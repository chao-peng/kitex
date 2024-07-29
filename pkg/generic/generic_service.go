/*
 * Copyright 2021 CloudWeGo Authors
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

package generic

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/gopkg/protocol/thrift/base"

	"github.com/cloudwego/kitex/internal/generic/proto"
	gthrift "github.com/cloudwego/kitex/internal/generic/thrift"
	codecProto "github.com/cloudwego/kitex/pkg/remote/codec/protobuf"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

// Service generic service interface
type Service interface {
	// GenericCall handle the generic call
	GenericCall(ctx context.Context, method string, request interface{}) (response interface{}, err error)
}

// ServiceInfoWithCodec create a generic ServiceInfo with CodecInfo
func ServiceInfoWithCodec(g Generic) *serviceinfo.ServiceInfo {
	return newServiceInfo(g.PayloadCodecType(), g.MessageReaderWriter(), g.IDLServiceName())
}

// Deprecated: it's not used by kitex anymore.
// ServiceInfo create a generic ServiceInfo
func ServiceInfo(pcType serviceinfo.PayloadCodec) *serviceinfo.ServiceInfo {
	return newServiceInfo(pcType, nil, "")
}

func newServiceInfo(pcType serviceinfo.PayloadCodec, messageReaderWriter interface{}, serviceName string) *serviceinfo.ServiceInfo {
	handlerType := (*Service)(nil)

	var methods map[string]serviceinfo.MethodInfo
	var svcName string

	if messageReaderWriter == nil {
		// note: binary generic cannot be used with multi-service feature
		svcName = serviceinfo.GenericService
		methods = map[string]serviceinfo.MethodInfo{
			serviceinfo.GenericMethod: serviceinfo.NewMethodInfo(callHandler, newGenericServiceCallArgs, newGenericServiceCallResult, false),
		}
	} else {
		svcName = serviceName
		methods = map[string]serviceinfo.MethodInfo{
			serviceinfo.GenericMethod: serviceinfo.NewMethodInfo(
				callHandler,
				func() interface{} { return &Args{inner: messageReaderWriter} },
				func() interface{} { return &Result{inner: messageReaderWriter} },
				false,
			),
		}
	}

	svcInfo := &serviceinfo.ServiceInfo{
		ServiceName:  svcName,
		HandlerType:  handlerType,
		Methods:      methods,
		PayloadCodec: pcType,
		Extra:        make(map[string]interface{}),
	}
	svcInfo.Extra["generic"] = true
	return svcInfo
}

func callHandler(ctx context.Context, handler, arg, result interface{}) error {
	realArg := arg.(*Args)
	realResult := result.(*Result)
	success, err := handler.(Service).GenericCall(ctx, realArg.Method, realArg.Request)
	if err != nil {
		return err
	}
	realResult.Success = success
	return nil
}

func newGenericServiceCallArgs() interface{} {
	return &Args{}
}

func newGenericServiceCallResult() interface{} {
	return &Result{}
}

// WithCodec set codec instance for Args or Result
type WithCodec interface {
	SetCodec(codec interface{})
}

// Args generic request
type Args struct {
	Request interface{}
	Method  string
	base    *base.Base
	inner   interface{}
}

var (
	_ codecProto.MessageWriterWithContext           = (*Args)(nil)
	_ codecProto.MessageReaderWithMethodWithContext = (*Args)(nil)
	_ WithCodec                                     = (*Args)(nil)
)

// SetCodec ...
func (g *Args) SetCodec(inner interface{}) {
	g.inner = inner
}

func (g *Args) GetOrSetBase() interface{} {
	if g.base == nil {
		g.base = base.NewBase()
	}
	return g.base
}

// Write ...
func (g *Args) Write(ctx context.Context, method string, bw *thrift.BinaryWriter, out io.Writer) error {
	if err, ok := g.inner.(error); ok {
		return err
	}
	if w, ok := g.inner.(gthrift.MessageWriter); ok {
		return w.Write(ctx, out, bw, g.Request, method, true, g.base)
	}
	return fmt.Errorf("unexpected Args writer type: %T", g.inner)
}

func (g *Args) WritePb(ctx context.Context, method string) (interface{}, error) {
	if err, ok := g.inner.(error); ok {
		return nil, err
	}
	if w, ok := g.inner.(proto.MessageWriter); ok {
		return w.Write(ctx, g.Request, method, true)
	}
	return nil, fmt.Errorf("unexpected Args writer type: %T", g.inner)
}

// Read ...
func (g *Args) Read(ctx context.Context, method string, dataLen int, in io.Reader) error {
	if err, ok := g.inner.(error); ok {
		return err
	}
	if rw, ok := g.inner.(gthrift.MessageReader); ok {
		g.Method = method
		var err error
		g.Request, err = rw.Read(ctx, method, false, dataLen, in)
		return err
	}
	return fmt.Errorf("unexpected Args reader type: %T", g.inner)
}

func (g *Args) ReadPb(ctx context.Context, method string, in []byte) error {
	if err, ok := g.inner.(error); ok {
		return err
	}
	if w, ok := g.inner.(proto.MessageReader); ok {
		g.Method = method
		var err error
		g.Request, err = w.Read(ctx, method, false, in)
		return err
	}
	return fmt.Errorf("unexpected Args reader type: %T", g.inner)
}

// GetFirstArgument implements util.KitexArgs.
func (g *Args) GetFirstArgument() interface{} {
	return g.Request
}

// Result generic response
type Result struct {
	Success interface{}
	inner   interface{}
}

var (
	_ codecProto.MessageWriterWithContext           = (*Result)(nil)
	_ codecProto.MessageReaderWithMethodWithContext = (*Result)(nil)
	_ WithCodec                                     = (*Result)(nil)
)

// SetCodec ...
func (r *Result) SetCodec(inner interface{}) {
	r.inner = inner
}

// Write ...
func (r *Result) Write(ctx context.Context, method string, bw *thrift.BinaryWriter, out io.Writer) error {
	if err, ok := r.inner.(error); ok {
		return err
	}
	if w, ok := r.inner.(gthrift.MessageWriter); ok {
		return w.Write(ctx, out, bw, r.Success, method, false, nil)
	}
	return fmt.Errorf("unexpected Result writer type: %T", r.inner)
}

func (r *Result) WritePb(ctx context.Context, method string) (interface{}, error) {
	if err, ok := r.inner.(error); ok {
		return nil, err
	}
	if w, ok := r.inner.(proto.MessageWriter); ok {
		return w.Write(ctx, r.Success, method, false)
	}
	return nil, fmt.Errorf("unexpected Result writer type: %T", r.inner)
}

// Read ...
func (r *Result) Read(ctx context.Context, method string, dataLen int, in io.Reader) error {
	if err, ok := r.inner.(error); ok {
		return err
	}
	if w, ok := r.inner.(gthrift.MessageReader); ok {
		var err error
		r.Success, err = w.Read(ctx, method, true, dataLen, in)
		return err
	}
	return fmt.Errorf("unexpected Result reader type: %T", r.inner)
}

func (r *Result) ReadPb(ctx context.Context, method string, in []byte) error {
	if err, ok := r.inner.(error); ok {
		return err
	}
	if w, ok := r.inner.(proto.MessageReader); ok {
		var err error
		r.Success, err = w.Read(ctx, method, true, in)
		return err
	}
	return fmt.Errorf("unexpected Result reader type: %T", r.inner)
}

// GetSuccess implements util.KitexResult.
func (r *Result) GetSuccess() interface{} {
	if !r.IsSetSuccess() {
		return nil
	}
	return r.Success
}

// SetSuccess implements util.KitexResult.
func (r *Result) SetSuccess(x interface{}) {
	r.Success = x
}

// IsSetSuccess ...
func (r *Result) IsSetSuccess() bool {
	return r.Success != nil
}

// GetResult ...
func (r *Result) GetResult() interface{} {
	return r.Success
}
