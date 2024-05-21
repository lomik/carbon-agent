// Code generated by protoc-gen-gogo.
// source: carbon.proto
// DO NOT EDIT!

/*
Package carbonpb is a generated protocol buffer package.

It is generated from these files:

	carbon.proto

It has these top-level messages:

	Point
	Metric
	Payload
	CacheRequest
*/
package carbonpb

import (
	fmt "fmt"

	proto "github.com/gogo/protobuf/proto"

	math "math"

	_ "github.com/gogo/protobuf/gogoproto"

	context "golang.org/x/net/context"

	grpc "google.golang.org/grpc"

	io "io"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Point struct {
	Timestamp uint32  `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Value     float64 `protobuf:"fixed64,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Point) Reset()                    { *m = Point{} }
func (m *Point) String() string            { return proto.CompactTextString(m) }
func (*Point) ProtoMessage()               {}
func (*Point) Descriptor() ([]byte, []int) { return fileDescriptorCarbon, []int{0} }

type Metric struct {
	Metric string  `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
	Points []Point `protobuf:"bytes,2,rep,name=points" json:"points"`
}

func (m *Metric) Reset()                    { *m = Metric{} }
func (m *Metric) String() string            { return proto.CompactTextString(m) }
func (*Metric) ProtoMessage()               {}
func (*Metric) Descriptor() ([]byte, []int) { return fileDescriptorCarbon, []int{1} }

func (m *Metric) GetPoints() []Point {
	if m != nil {
		return m.Points
	}
	return nil
}

type Payload struct {
	Metrics []*Metric `protobuf:"bytes,1,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *Payload) Reset()                    { *m = Payload{} }
func (m *Payload) String() string            { return proto.CompactTextString(m) }
func (*Payload) ProtoMessage()               {}
func (*Payload) Descriptor() ([]byte, []int) { return fileDescriptorCarbon, []int{2} }

func (m *Payload) GetMetrics() []*Metric {
	if m != nil {
		return m.Metrics
	}
	return nil
}

type CacheRequest struct {
	Metrics []string `protobuf:"bytes,1,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *CacheRequest) Reset()                    { *m = CacheRequest{} }
func (m *CacheRequest) String() string            { return proto.CompactTextString(m) }
func (*CacheRequest) ProtoMessage()               {}
func (*CacheRequest) Descriptor() ([]byte, []int) { return fileDescriptorCarbon, []int{3} }

func init() {
	proto.RegisterType((*Point)(nil), "carbonpb.Point")
	proto.RegisterType((*Metric)(nil), "carbonpb.Metric")
	proto.RegisterType((*Payload)(nil), "carbonpb.Payload")
	proto.RegisterType((*CacheRequest)(nil), "carbonpb.CacheRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Carbon service

type CarbonClient interface {
	// Same as carbonlink
	CacheQuery(ctx context.Context, in *CacheRequest, opts ...grpc.CallOption) (*Payload, error)
}

type carbonClient struct {
	cc *grpc.ClientConn
}

func NewCarbonClient(cc *grpc.ClientConn) CarbonClient {
	return &carbonClient{cc}
}

func (c *carbonClient) CacheQuery(ctx context.Context, in *CacheRequest, opts ...grpc.CallOption) (*Payload, error) {
	out := new(Payload)
	err := grpc.Invoke(ctx, "/carbonpb.Carbon/CacheQuery", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Carbon service

type CarbonServer interface {
	// Same as carbonlink
	CacheQuery(context.Context, *CacheRequest) (*Payload, error)
}

func RegisterCarbonServer(s *grpc.Server, srv CarbonServer) {
	s.RegisterService(&_Carbon_serviceDesc, srv)
}

func _Carbon_CacheQuery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CacheRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CarbonServer).CacheQuery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/carbonpb.Carbon/CacheQuery",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CarbonServer).CacheQuery(ctx, req.(*CacheRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Carbon_serviceDesc = grpc.ServiceDesc{
	ServiceName: "carbonpb.Carbon",
	HandlerType: (*CarbonServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CacheQuery",
			Handler:    _Carbon_CacheQuery_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptorCarbon,
}

func (m *Point) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Point) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Timestamp != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintCarbon(dAtA, i, uint64(m.Timestamp))
	}
	if m.Value != 0 {
		dAtA[i] = 0x11
		i++
		i = encodeFixed64Carbon(dAtA, i, uint64(math.Float64bits(float64(m.Value))))
	}
	return i, nil
}

func (m *Metric) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metric) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Metric) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintCarbon(dAtA, i, uint64(len(m.Metric)))
		i += copy(dAtA[i:], m.Metric)
	}
	if len(m.Points) > 0 {
		for _, msg := range m.Points {
			dAtA[i] = 0x12
			i++
			i = encodeVarintCarbon(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *Payload) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Payload) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, msg := range m.Metrics {
			dAtA[i] = 0xa
			i++
			i = encodeVarintCarbon(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *CacheRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CacheRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, s := range m.Metrics {
			dAtA[i] = 0xa
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	return i, nil
}

func encodeFixed64Carbon(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Carbon(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintCarbon(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Point) Size() (n int) {
	var l int
	_ = l
	if m.Timestamp != 0 {
		n += 1 + sovCarbon(uint64(m.Timestamp))
	}
	if m.Value != 0 {
		n += 9
	}
	return n
}

func (m *Metric) Size() (n int) {
	var l int
	_ = l
	l = len(m.Metric)
	if l > 0 {
		n += 1 + l + sovCarbon(uint64(l))
	}
	if len(m.Points) > 0 {
		for _, e := range m.Points {
			l = e.Size()
			n += 1 + l + sovCarbon(uint64(l))
		}
	}
	return n
}

func (m *Payload) Size() (n int) {
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, e := range m.Metrics {
			l = e.Size()
			n += 1 + l + sovCarbon(uint64(l))
		}
	}
	return n
}

func (m *CacheRequest) Size() (n int) {
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, s := range m.Metrics {
			l = len(s)
			n += 1 + l + sovCarbon(uint64(l))
		}
	}
	return n
}

func sovCarbon(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCarbon(x uint64) (n int) {
	return sovCarbon(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Point) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Point: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Point: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += 8
			v = uint64(dAtA[iNdEx-8])
			v |= uint64(dAtA[iNdEx-7]) << 8
			v |= uint64(dAtA[iNdEx-6]) << 16
			v |= uint64(dAtA[iNdEx-5]) << 24
			v |= uint64(dAtA[iNdEx-4]) << 32
			v |= uint64(dAtA[iNdEx-3]) << 40
			v |= uint64(dAtA[iNdEx-2]) << 48
			v |= uint64(dAtA[iNdEx-1]) << 56
			m.Value = float64(math.Float64frombits(v))
		default:
			iNdEx = preIndex
			skippy, err := skipCarbon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCarbon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Metric) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Metric: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metric: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metric", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metric = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Points", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCarbon
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Points = append(m.Points, Point{})
			if err := m.Points[len(m.Points)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCarbon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Payload) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Payload: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Payload: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metrics", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCarbon
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, &Metric{})
			if err := m.Metrics[len(m.Metrics)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCarbon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CacheRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CacheRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CacheRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metrics", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCarbon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCarbon(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCarbon
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCarbon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthCarbon
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCarbon
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipCarbon(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthCarbon = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCarbon   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("carbon.proto", fileDescriptorCarbon) }

var fileDescriptorCarbon = []byte{
	// 278 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0x90, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x86, 0x97, 0xe9, 0x3a, 0x7b, 0xac, 0xa8, 0x41, 0x46, 0x19, 0x52, 0x4b, 0xaf, 0x8a, 0xb0,
	0x0e, 0x26, 0x5e, 0x88, 0x77, 0xeb, 0xb5, 0x38, 0xf3, 0x06, 0x49, 0x8d, 0x5d, 0x61, 0x5d, 0x6a,
	0x9b, 0x08, 0x7b, 0xc3, 0x5d, 0xfa, 0x04, 0x22, 0x7d, 0x12, 0xf1, 0x64, 0xb3, 0xf3, 0xee, 0xfc,
	0xe1, 0xff, 0xbe, 0x73, 0x08, 0x78, 0x19, 0xaf, 0x85, 0x5a, 0x27, 0x55, 0xad, 0xb4, 0xa2, 0x27,
	0x36, 0x55, 0x62, 0x3c, 0xc9, 0x0b, 0xbd, 0x34, 0x22, 0xc9, 0x54, 0x39, 0xcd, 0x55, 0xae, 0xa6,
	0x58, 0x10, 0xe6, 0x0d, 0x13, 0x06, 0x9c, 0x2c, 0x18, 0x3d, 0xc2, 0x60, 0xa1, 0x8a, 0xb5, 0xa6,
	0xd7, 0xe0, 0xea, 0xa2, 0x94, 0x8d, 0xe6, 0x65, 0xe5, 0x93, 0x90, 0xc4, 0x67, 0xac, 0x7b, 0xa0,
	0x57, 0x30, 0xf8, 0xe0, 0x2b, 0x23, 0xfd, 0x7e, 0x48, 0x62, 0xc2, 0x6c, 0x88, 0x9e, 0xc1, 0x79,
	0x92, 0xba, 0x2e, 0x32, 0x3a, 0x02, 0xa7, 0xc4, 0x09, 0x51, 0x97, 0xed, 0x12, 0x9d, 0x80, 0x53,
	0xfd, 0xea, 0x1b, 0xbf, 0x1f, 0x1e, 0xc5, 0xa7, 0xb3, 0xf3, 0x64, 0x7f, 0x68, 0x82, 0x6b, 0xe7,
	0xc7, 0xdb, 0xaf, 0x9b, 0x1e, 0xdb, 0x95, 0xa2, 0x7b, 0x18, 0x2e, 0xf8, 0x66, 0xa5, 0xf8, 0x2b,
	0xbd, 0x85, 0xa1, 0x75, 0x34, 0x3e, 0x41, 0xf4, 0xa2, 0x43, 0xed, 0x52, 0xb6, 0x2f, 0x44, 0x31,
	0x78, 0x29, 0xcf, 0x96, 0x92, 0xc9, 0x77, 0x23, 0x1b, 0x4d, 0xfd, 0xff, 0xac, 0xfb, 0xd7, 0x9c,
	0xa5, 0xe0, 0xa4, 0x68, 0xa1, 0x0f, 0x00, 0xc8, 0xbc, 0x18, 0x59, 0x6f, 0xe8, 0xa8, 0x93, 0x1f,
	0x9a, 0xc6, 0x97, 0x07, 0xf7, 0xda, 0xc3, 0xa2, 0xde, 0xdc, 0xdb, 0xb6, 0x01, 0xf9, 0x6c, 0x03,
	0xf2, 0xdd, 0x06, 0x44, 0x38, 0xf8, 0x91, 0x77, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x96, 0x97,
	0x56, 0x0a, 0x91, 0x01, 0x00, 0x00,
}
