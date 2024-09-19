/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ttstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/lang/mcache"
	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/pkg/streamx"
	"github.com/cloudwego/kitex/pkg/streamx/provider/ttstream/container"
	"github.com/cloudwego/netpoll"
)

const (
	clientTransport int32 = 1
	serverTransport int32 = 2

	streamCacheSize = 32
	frameChanSize   = 32
)

var _ Object = (*transport)(nil)

type transport struct {
	kind     int32
	sinfo    *serviceinfo.ServiceInfo
	conn     netpoll.Connection
	streams  sync.Map                 // key=streamID val=streamIO
	scache   []*stream                // size is streamCacheSize
	spipe    *container.Pipe[*stream] // in-coming stream channel
	wchannel chan *Frame

	// for scavenger check
	lastActive atomic.Value // time.Time
}

func newTransport(kind int32, sinfo *serviceinfo.ServiceInfo, conn netpoll.Connection) *transport {
	_ = conn.SetDeadline(time.Now().Add(time.Hour))
	t := &transport{
		kind:     kind,
		sinfo:    sinfo,
		conn:     conn,
		streams:  sync.Map{},
		spipe:    container.NewPipe[*stream](),
		scache:   make([]*stream, 0, streamCacheSize),
		wchannel: make(chan *Frame, frameChanSize),
	}
	go func() {
		err := t.loopRead()
		if err != nil && errors.Is(err, io.EOF) {
			klog.Warnf("trans loop read err: %v", err)
		}
	}()
	go func() {
		err := t.loopWrite()
		if err != nil && errors.Is(err, io.EOF) {
			klog.Warnf("trans loop write err: %v", err)
		}
	}()
	return t
}

func (t *transport) storeStreamIO(ctx context.Context, s *stream) {
	t.streams.Store(s.sid, newStreamIO(ctx, s))
}

func (t *transport) loadStreamIO(sid int32) (sio *streamIO, ok bool) {
	val, ok := t.streams.Load(sid)
	if !ok {
		return sio, false
	}
	sio = val.(*streamIO)
	return sio, true
}

func (t *transport) loopRead() error {
	reader := newReaderBuffer(t.conn.Reader())
	for {
		now := time.Now()
		t.lastActive.Store(now)

		// decode frame
		fr, err := DecodeFrame(context.Background(), reader)
		if err != nil {
			return err
		}
		klog.Debugf("transport[%d] DecodeFrame: fr=%v", t.kind, fr)

		switch fr.typ {
		case metaFrameType:
			sio, ok := t.loadStreamIO(fr.sid)
			if !ok {
				klog.Errorf("transport[%d] read a unknown stream meta: sid=%d", t.kind, fr.sid)
				continue
			}
			err = sio.stream.readMetaFrame(fr.meta, fr.header, fr.payload)
			if err != nil {
				return err
			}
		case headerFrameType:
			switch t.kind {
			case serverTransport:
				// Header Frame: server recv a new stream
				smode := t.sinfo.MethodInfo(fr.method).StreamingMode()
				s := newStream(context.Background(), t, smode, fr.streamFrame)
				klog.Debugf("transport[%d] read a new stream: sid=%d", t.kind, s.sid)
				t.spipe.Write(s)
			case clientTransport:
				// Header Frame: client recv header
				sio, ok := t.loadStreamIO(fr.sid)
				if !ok {
					klog.Errorf("transport[%d] read a unknown stream header: sid=%d", t.kind, fr.sid)
					continue
				}
				err = sio.stream.readHeader(fr.header)
				if err != nil {
					return err
				}
			}
		case dataFrameType:
			// Data Frame: decode and distribute data
			sio, ok := t.loadStreamIO(fr.sid)
			if !ok {
				klog.Errorf("transport[%d] read a unknown stream data: sid=%d", t.kind, fr.sid)
				continue
			}
			sio.input(fr)
		case trailerFrameType:
			// Trailer Frame: recv trailer, Close read direction
			sio, ok := t.loadStreamIO(fr.sid)
			if !ok {
				klog.Errorf("transport[%d] read a unknown stream trailer: sid=%d", t.kind, fr.sid)
				continue
			}
			if err = sio.stream.readTrailer(fr.trailer); err != nil {
				return err
			}
		}
	}
}

func (t *transport) loopWrite() (err error) {
	writer := newWriterBuffer(t.conn.Writer())
	for {
		select {
		case frame, ok := <-t.wchannel:
			if !ok {
				// closed
				return nil
			}
			if err = EncodeFrame(context.Background(), writer, frame); err != nil {
				return err
			}
			if len(t.wchannel) == 0 {
				if err = t.conn.Writer().Flush(); err != nil {
					return err
				}
			}
		}
	}
}

// writeFrame is concurrent safe
func (t *transport) writeFrame(meta streamFrame, ftype int32, data any) (err error) {
	var payload []byte
	if data != nil {
		// payload should be written nocopy
		payload, err = EncodePayload(context.Background(), data)
		if err != nil {
			return err
		}
	}
	frame := newFrame(meta, ftype, payload)
	t.wchannel <- frame
	return nil
}

func (t *transport) Available() bool {
	v := t.lastActive.Load()
	if v == nil {
		return true
	}
	lastActive := v.(time.Time)
	// TODO: let unavailable time configurable
	return time.Now().Sub(lastActive) < time.Minute*10
}

func (t *transport) Close() (err error) {
	klog.Warnf("transport[%s] is closing", t.conn.LocalAddr())
	t.spipe.Close()
	close(t.wchannel)
	err = t.conn.Close()
	return err
}

func (t *transport) streamSend(sid int32, method string, wheader streamx.Header, res any) (err error) {
	if len(wheader) > 0 {
		err = t.streamSendHeader(sid, method, wheader)
		if err != nil {
			return err
		}
	}
	return t.writeFrame(streamFrame{sid: sid, method: method}, dataFrameType, res)
}

func (t *transport) streamSendHeader(sid int32, method string, header streamx.Header) (err error) {
	return t.writeFrame(streamFrame{sid: sid, method: method, header: header}, headerFrameType, nil)
}

func (t *transport) streamSendTrailer(sid int32, method string, trailer streamx.Trailer) (err error) {
	return t.writeFrame(streamFrame{sid: sid, method: method, trailer: trailer}, trailerFrameType, nil)
}

func (t *transport) streamRecv(sid int32, data any) (err error) {
	sio, ok := t.loadStreamIO(sid)
	if !ok {
		return io.EOF
	}
	f, err := sio.output()
	if err != nil {
		return err
	}
	err = DecodePayload(context.Background(), f.payload, data.(thrift.FastCodec))
	// payload will not be access after decode
	mcache.Free(f.payload)
	recycleFrame(f)
	return nil
}

func (t *transport) streamCloseRecv(s *stream) (err error) {
	sio, ok := t.loadStreamIO(s.sid)
	if !ok {
		return fmt.Errorf("stream not found in stream map: sid=%d", s.sid)
	}
	sio.eof()
	return nil
}

func (t *transport) streamCancel(s *stream) (err error) {
	sio, ok := t.loadStreamIO(s.sid)
	if !ok {
		return fmt.Errorf("stream not found in stream map: sid=%d", s.sid)
	}
	sio.cancel()
	return nil
}

func (t *transport) streamClose(s *stream) (err error) {
	// remove stream from transport
	t.streams.Delete(s.sid)
	return nil
}

var clientStreamID int32

// newStream create new stream on current connection
// it's typically used by client side
// newStream is concurrency safe
func (t *transport) newStream(
	ctx context.Context, method string, header streamx.Header) (*stream, error) {
	if t.kind != clientTransport {
		return nil, fmt.Errorf("transport already be used as other kind")
	}
	sid := atomic.AddInt32(&clientStreamID, 1)
	smode := t.sinfo.MethodInfo(method).StreamingMode()
	smeta := streamFrame{
		sid:    sid,
		method: method,
		header: header,
	}
	s := newStream(ctx, t, smode, smeta)
	err := t.writeFrame(smeta, headerFrameType, nil) // create stream
	if err != nil {
		return nil, err
	}
	return s, nil
}

// readStream wait for a new incoming stream on current connection
// it's typically used by server side
func (t *transport) readStream() (*stream, error) {
	if t.kind != serverTransport {
		return nil, fmt.Errorf("transport already be used as other kind")
	}
READ:
	if len(t.scache) > 0 {
		s := t.scache[len(t.scache)-1]
		t.scache = t.scache[:len(t.scache)-1]
		return s, nil
	}
	n, err := t.spipe.Read(t.scache[0:streamCacheSize])
	if err != nil {
		if errors.Is(err, container.ErrPipeEOF) {
			return nil, io.EOF
		}
		return nil, err
	}
	if n == 0 {
		panic("Assert: N == 0 !")
	}
	t.scache = t.scache[:n]
	goto READ
}
