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

package ttstream_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudwego/kitex/pkg/protocol/bthrift"

	"github.com/apache/thrift/lib/go/thrift"
	kutils "github.com/cloudwego/kitex/pkg/utils"
)

// unused protection
var (
	_ = fmt.Formatter(nil)
	_ = (*bytes.Buffer)(nil)
	_ = (*strings.Builder)(nil)
	_ = reflect.Type(nil)
	_ = thrift.TProtocol(nil)
	_ = bthrift.BinaryWriter(nil)
)

var fieldIDToName_Request = map[int16]string{
	1: "Type",
	2: "Message",
}

var fieldIDToName_Response = map[int16]string{
	1: "Type",
	2: "Message",
}

type Request struct {
	Type    int32  `thrift:"Type,1" frugal:"1,default,i32" json:"Type"`
	Message string `thrift:"Message,2" frugal:"2,default,string" json:"Message"`
}

type Response struct {
	Type    int32  `thrift:"Type,1" frugal:"1,default,i32" json:"Type"`
	Message string `thrift:"Message,2" frugal:"2,default,string" json:"Message"`
}

type ServerPingPongArgs struct {
	Req *Request `thrift:"req,1" frugal:"1,default,Request" json:"req"`
}

type ServerPingPongResult struct {
	Success *Response `thrift:"success,0,optional" frugal:"0,optional,Response" json:"success,omitempty"`
}

func NewServerPingPongArgs() *ServerPingPongArgs {
	return &ServerPingPongArgs{}
}

func NewServerPingPongResult() *ServerPingPongResult {
	return &ServerPingPongResult{}
}

func (p ServerPingPongResult) GetSuccess() *Response {
	return p.Success
}

func (p *Request) FastRead(buf []byte) (int, error) {
	var err error
	var offset int
	var l int
	var fieldTypeId thrift.TType
	var fieldId int16
	_, l, err = bthrift.Binary.ReadStructBegin(buf)
	offset += l
	if err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, l, err = bthrift.Binary.ReadFieldBegin(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.I32 {
				l, err = p.FastReadField1(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 2:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField2(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		default:
			l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
			offset += l
			if err != nil {
				goto SkipFieldError
			}
		}

		l, err = bthrift.Binary.ReadFieldEnd(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldEndError
		}
	}
	l, err = bthrift.Binary.ReadStructEnd(buf[offset:])
	offset += l
	if err != nil {
		goto ReadStructEndError
	}

	return offset, nil
ReadStructBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_Request[fieldId]), err)
SkipFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
ReadFieldEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
}

func (p *Request) FastReadField1(buf []byte) (int, error) {
	offset := 0

	var _field int32
	if v, l, err := bthrift.Binary.ReadI32(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		_field = v

	}
	p.Type = _field
	return offset, nil
}

func (p *Request) FastReadField2(buf []byte) (int, error) {
	offset := 0

	var _field string
	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		_field = v

	}
	p.Message = _field
	return offset, nil
}

// for compatibility
func (p *Request) FastWrite(buf []byte) int {
	return 0
}

func (p *Request) FastWriteNocopy(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteStructBegin(buf[offset:], "Request")
	if p != nil {
		offset += p.fastWriteField1(buf[offset:], binaryWriter)
		offset += p.fastWriteField2(buf[offset:], binaryWriter)
	}
	offset += bthrift.Binary.WriteFieldStop(buf[offset:])
	offset += bthrift.Binary.WriteStructEnd(buf[offset:])
	return offset
}

func (p *Request) BLength() int {
	l := 0
	l += bthrift.Binary.StructBeginLength("Request")
	if p != nil {
		l += p.field1Length()
		l += p.field2Length()
	}
	l += bthrift.Binary.FieldStopLength()
	l += bthrift.Binary.StructEndLength()
	return l
}

func (p *Request) fastWriteField1(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Type", thrift.I32, 1)
	offset += bthrift.Binary.WriteI32(buf[offset:], p.Type)
	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Request) fastWriteField2(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Message", thrift.STRING, 2)
	offset += bthrift.Binary.WriteStringNocopy(buf[offset:], binaryWriter, p.Message)
	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Request) field1Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Type", thrift.I32, 1)
	l += bthrift.Binary.I32Length(p.Type)
	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Request) field2Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Message", thrift.STRING, 2)
	l += bthrift.Binary.StringLengthNocopy(p.Message)
	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Request) DeepCopy(s interface{}) error {
	src, ok := s.(*Request)
	if !ok {
		return fmt.Errorf("%T's type not matched %T", s, p)
	}

	p.Type = src.Type

	if src.Message != "" {
		p.Message = kutils.StringDeepCopy(src.Message)
	}

	return nil
}

func (p *Response) FastRead(buf []byte) (int, error) {
	var err error
	var offset int
	var l int
	var fieldTypeId thrift.TType
	var fieldId int16
	_, l, err = bthrift.Binary.ReadStructBegin(buf)
	offset += l
	if err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, l, err = bthrift.Binary.ReadFieldBegin(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.I32 {
				l, err = p.FastReadField1(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 2:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField2(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		default:
			l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
			offset += l
			if err != nil {
				goto SkipFieldError
			}
		}

		l, err = bthrift.Binary.ReadFieldEnd(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldEndError
		}
	}
	l, err = bthrift.Binary.ReadStructEnd(buf[offset:])
	offset += l
	if err != nil {
		goto ReadStructEndError
	}

	return offset, nil
ReadStructBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_Response[fieldId]), err)
SkipFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
ReadFieldEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
}

func (p *Response) FastReadField1(buf []byte) (int, error) {
	offset := 0

	var _field int32
	if v, l, err := bthrift.Binary.ReadI32(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		_field = v

	}
	p.Type = _field
	return offset, nil
}

func (p *Response) FastReadField2(buf []byte) (int, error) {
	offset := 0

	var _field string
	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		_field = v

	}
	p.Message = _field
	return offset, nil
}

// for compatibility
func (p *Response) FastWrite(buf []byte) int {
	return 0
}

func (p *Response) FastWriteNocopy(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteStructBegin(buf[offset:], "Response")
	if p != nil {
		offset += p.fastWriteField1(buf[offset:], binaryWriter)
		offset += p.fastWriteField2(buf[offset:], binaryWriter)
	}
	offset += bthrift.Binary.WriteFieldStop(buf[offset:])
	offset += bthrift.Binary.WriteStructEnd(buf[offset:])
	return offset
}

func (p *Response) BLength() int {
	l := 0
	l += bthrift.Binary.StructBeginLength("Response")
	if p != nil {
		l += p.field1Length()
		l += p.field2Length()
	}
	l += bthrift.Binary.FieldStopLength()
	l += bthrift.Binary.StructEndLength()
	return l
}

func (p *Response) fastWriteField1(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Type", thrift.I32, 1)
	offset += bthrift.Binary.WriteI32(buf[offset:], p.Type)
	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Response) fastWriteField2(buf []byte, binaryWriter bthrift.BinaryWriter) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Message", thrift.STRING, 2)
	offset += bthrift.Binary.WriteStringNocopy(buf[offset:], binaryWriter, p.Message)
	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Response) field1Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Type", thrift.I32, 1)
	l += bthrift.Binary.I32Length(p.Type)
	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Response) field2Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Message", thrift.STRING, 2)
	l += bthrift.Binary.StringLengthNocopy(p.Message)
	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Response) DeepCopy(s interface{}) error {
	src, ok := s.(*Response)
	if !ok {
		return fmt.Errorf("%T's type not matched %T", s, p)
	}

	p.Type = src.Type

	if src.Message != "" {
		p.Message = kutils.StringDeepCopy(src.Message)
	}

	return nil
}
