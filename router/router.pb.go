// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: router.proto

package router

import (
	context "context"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Header struct {
	ID              string            `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	OrgActorID      string            `protobuf:"bytes,2,opt,name=OrgActorID,proto3" json:"OrgActorID,omitempty"`
	OrgActorType    string            `protobuf:"bytes,3,opt,name=OrgActorType,proto3" json:"OrgActorType,omitempty"`
	PrevActorType   string            `protobuf:"bytes,4,opt,name=PrevActorType,proto3" json:"PrevActorType,omitempty"`
	TargetActorID   string            `protobuf:"bytes,7,opt,name=TargetActorID,proto3" json:"TargetActorID,omitempty"`
	TargetActorType string            `protobuf:"bytes,8,opt,name=TargetActorType,proto3" json:"TargetActorType,omitempty"`
	Event           string            `protobuf:"bytes,9,opt,name=Event,proto3" json:"Event,omitempty"`
	Token           string            `protobuf:"bytes,10,opt,name=Token,proto3" json:"Token,omitempty"`
	Timestamp       int64             `protobuf:"varint,11,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	Custom          map[string]string `protobuf:"bytes,12,rep,name=Custom,proto3" json:"Custom,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *Header) Reset()         { *m = Header{} }
func (m *Header) String() string { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()    {}
func (*Header) Descriptor() ([]byte, []int) {
	return fileDescriptor_367072455c71aedc, []int{0}
}
func (m *Header) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Header) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Header.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Header) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Header.Merge(m, src)
}
func (m *Header) XXX_Size() int {
	return m.Size()
}
func (m *Header) XXX_DiscardUnknown() {
	xxx_messageInfo_Header.DiscardUnknown(m)
}

var xxx_messageInfo_Header proto.InternalMessageInfo

func (m *Header) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Header) GetOrgActorID() string {
	if m != nil {
		return m.OrgActorID
	}
	return ""
}

func (m *Header) GetOrgActorType() string {
	if m != nil {
		return m.OrgActorType
	}
	return ""
}

func (m *Header) GetPrevActorType() string {
	if m != nil {
		return m.PrevActorType
	}
	return ""
}

func (m *Header) GetTargetActorID() string {
	if m != nil {
		return m.TargetActorID
	}
	return ""
}

func (m *Header) GetTargetActorType() string {
	if m != nil {
		return m.TargetActorType
	}
	return ""
}

func (m *Header) GetEvent() string {
	if m != nil {
		return m.Event
	}
	return ""
}

func (m *Header) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *Header) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Header) GetCustom() map[string]string {
	if m != nil {
		return m.Custom
	}
	return nil
}

type Message struct {
	Header *Header `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Body   []byte  `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_367072455c71aedc, []int{1}
}
func (m *Message) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Message.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return m.Size()
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Message) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type RouteReq struct {
	Msg *Message `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (m *RouteReq) Reset()         { *m = RouteReq{} }
func (m *RouteReq) String() string { return proto.CompactTextString(m) }
func (*RouteReq) ProtoMessage()    {}
func (*RouteReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_367072455c71aedc, []int{2}
}
func (m *RouteReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RouteReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RouteReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RouteReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RouteReq.Merge(m, src)
}
func (m *RouteReq) XXX_Size() int {
	return m.Size()
}
func (m *RouteReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RouteReq.DiscardUnknown(m)
}

var xxx_messageInfo_RouteReq proto.InternalMessageInfo

func (m *RouteReq) GetMsg() *Message {
	if m != nil {
		return m.Msg
	}
	return nil
}

type RouteRes struct {
	Msg *Message `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (m *RouteRes) Reset()         { *m = RouteRes{} }
func (m *RouteRes) String() string { return proto.CompactTextString(m) }
func (*RouteRes) ProtoMessage()    {}
func (*RouteRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_367072455c71aedc, []int{3}
}
func (m *RouteRes) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RouteRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RouteRes.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RouteRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RouteRes.Merge(m, src)
}
func (m *RouteRes) XXX_Size() int {
	return m.Size()
}
func (m *RouteRes) XXX_DiscardUnknown() {
	xxx_messageInfo_RouteRes.DiscardUnknown(m)
}

var xxx_messageInfo_RouteRes proto.InternalMessageInfo

func (m *RouteRes) GetMsg() *Message {
	if m != nil {
		return m.Msg
	}
	return nil
}

func init() {
	proto.RegisterType((*Header)(nil), "router.Header")
	proto.RegisterMapType((map[string]string)(nil), "router.Header.CustomEntry")
	proto.RegisterType((*Message)(nil), "router.Message")
	proto.RegisterType((*RouteReq)(nil), "router.routeReq")
	proto.RegisterType((*RouteRes)(nil), "router.routeRes")
}

func init() { proto.RegisterFile("router.proto", fileDescriptor_367072455c71aedc) }

var fileDescriptor_367072455c71aedc = []byte{
	// 388 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xcd, 0x8a, 0xda, 0x50,
	0x14, 0xc7, 0xf3, 0x61, 0xa3, 0x9e, 0xa4, 0x2a, 0x87, 0x2e, 0x2e, 0x52, 0x82, 0x0d, 0xa5, 0x64,
	0x53, 0x0b, 0xe9, 0xa6, 0x1f, 0x2b, 0x5b, 0x85, 0xba, 0x28, 0x2d, 0x21, 0x2f, 0x10, 0xf5, 0x92,
	0x11, 0x27, 0xb9, 0x99, 0x9b, 0xab, 0x90, 0x67, 0x98, 0xcd, 0x3c, 0xd6, 0x2c, 0x5d, 0xce, 0x72,
	0xd0, 0x17, 0x19, 0x72, 0x93, 0x60, 0x22, 0xcc, 0xee, 0xfe, 0x7f, 0xe7, 0xe7, 0xf9, 0xcb, 0x21,
	0x60, 0x71, 0xb6, 0x17, 0x94, 0x4f, 0x53, 0xce, 0x04, 0x43, 0xa3, 0x4c, 0xce, 0xbd, 0x0e, 0xc6,
	0x1f, 0x1a, 0x6e, 0x28, 0xc7, 0x01, 0x68, 0xcb, 0x39, 0x51, 0x27, 0xaa, 0xdb, 0xf7, 0xb5, 0xe5,
	0x1c, 0x6d, 0x80, 0x7f, 0x3c, 0x9a, 0xad, 0x05, 0xe3, 0xcb, 0x39, 0xd1, 0x24, 0x6f, 0x10, 0x74,
	0xc0, 0xaa, 0x53, 0x90, 0xa7, 0x94, 0xe8, 0xd2, 0x68, 0x31, 0xfc, 0x08, 0x6f, 0xff, 0x73, 0x7a,
	0xb8, 0x48, 0x1d, 0x29, 0xb5, 0x61, 0x61, 0x05, 0x21, 0x8f, 0xa8, 0xa8, 0xcb, 0xba, 0xa5, 0xd5,
	0x82, 0xe8, 0xc2, 0xb0, 0x01, 0xe4, 0xb6, 0x9e, 0xf4, 0xae, 0x31, 0xbe, 0x83, 0x37, 0x8b, 0x03,
	0x4d, 0x04, 0xe9, 0xcb, 0x79, 0x19, 0x0a, 0x1a, 0xb0, 0x1d, 0x4d, 0x08, 0x94, 0x54, 0x06, 0x7c,
	0x0f, 0xfd, 0x60, 0x1b, 0xd3, 0x4c, 0x84, 0x71, 0x4a, 0xcc, 0x89, 0xea, 0xea, 0xfe, 0x05, 0xa0,
	0x07, 0xc6, 0xef, 0x7d, 0x26, 0x58, 0x4c, 0xac, 0x89, 0xee, 0x9a, 0xde, 0x78, 0x5a, 0x5d, 0xb1,
	0xbc, 0xd9, 0xb4, 0x1c, 0x2e, 0x12, 0xc1, 0x73, 0xbf, 0x32, 0xc7, 0xdf, 0xc1, 0x6c, 0x60, 0x1c,
	0x81, 0xbe, 0xa3, 0x79, 0x75, 0xd7, 0xe2, 0x59, 0xfc, 0x91, 0x43, 0x78, 0xbb, 0xa7, 0xd5, 0x4d,
	0xcb, 0xf0, 0x43, 0xfb, 0xa6, 0x3a, 0x0b, 0xe8, 0xfe, 0xa5, 0x59, 0x16, 0x46, 0x14, 0x3f, 0x81,
	0x71, 0x23, 0x3b, 0xe4, 0x2f, 0x4d, 0x6f, 0xd0, 0x6e, 0xf6, 0xab, 0x29, 0x22, 0x74, 0x56, 0x6c,
	0x93, 0xcb, 0x5d, 0x96, 0x2f, 0xdf, 0xce, 0x67, 0xe8, 0x49, 0xd9, 0xa7, 0x77, 0xf8, 0x01, 0xf4,
	0x38, 0x8b, 0xaa, 0x25, 0xc3, 0x7a, 0x49, 0xd5, 0xe2, 0x17, 0xb3, 0x86, 0x9e, 0xd5, 0xba, 0xf6,
	0xba, 0xee, 0xfd, 0x84, 0xde, 0x6c, 0xbd, 0xa6, 0xa9, 0x60, 0x1c, 0xbf, 0x40, 0xb7, 0x50, 0xb6,
	0x49, 0x84, 0xa3, 0x5a, 0xae, 0xab, 0xc7, 0xd7, 0x24, 0x73, 0x94, 0x5f, 0xe4, 0xf1, 0x64, 0xab,
	0xc7, 0x93, 0xad, 0x3e, 0x9f, 0x6c, 0xf5, 0xe1, 0x6c, 0x2b, 0xc7, 0xb3, 0xad, 0x3c, 0x9d, 0x6d,
	0x65, 0x65, 0xc8, 0x0f, 0xf3, 0xeb, 0x4b, 0x00, 0x00, 0x00, 0xff, 0xff, 0x13, 0x77, 0xa6, 0xae,
	0xa8, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AcceptorClient is the client API for Acceptor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AcceptorClient interface {
	Routing(ctx context.Context, in *RouteReq, opts ...grpc.CallOption) (*RouteRes, error)
}

type acceptorClient struct {
	cc *grpc.ClientConn
}

func NewAcceptorClient(cc *grpc.ClientConn) AcceptorClient {
	return &acceptorClient{cc}
}

func (c *acceptorClient) Routing(ctx context.Context, in *RouteReq, opts ...grpc.CallOption) (*RouteRes, error) {
	out := new(RouteRes)
	err := c.cc.Invoke(ctx, "/router.Acceptor/routing", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AcceptorServer is the server API for Acceptor service.
type AcceptorServer interface {
	Routing(context.Context, *RouteReq) (*RouteRes, error)
}

// UnimplementedAcceptorServer can be embedded to have forward compatible implementations.
type UnimplementedAcceptorServer struct {
}

func (*UnimplementedAcceptorServer) Routing(ctx context.Context, req *RouteReq) (*RouteRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Routing not implemented")
}

func RegisterAcceptorServer(s *grpc.Server, srv AcceptorServer) {
	s.RegisterService(&_Acceptor_serviceDesc, srv)
}

func _Acceptor_Routing_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RouteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AcceptorServer).Routing(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/router.Acceptor/Routing",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AcceptorServer).Routing(ctx, req.(*RouteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Acceptor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "router.Acceptor",
	HandlerType: (*AcceptorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "routing",
			Handler:    _Acceptor_Routing_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "router.proto",
}

func (m *Header) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Header) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Header) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Custom) > 0 {
		for k := range m.Custom {
			v := m.Custom[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintRouter(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintRouter(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintRouter(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x62
		}
	}
	if m.Timestamp != 0 {
		i = encodeVarintRouter(dAtA, i, uint64(m.Timestamp))
		i--
		dAtA[i] = 0x58
	}
	if len(m.Token) > 0 {
		i -= len(m.Token)
		copy(dAtA[i:], m.Token)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.Token)))
		i--
		dAtA[i] = 0x52
	}
	if len(m.Event) > 0 {
		i -= len(m.Event)
		copy(dAtA[i:], m.Event)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.Event)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.TargetActorType) > 0 {
		i -= len(m.TargetActorType)
		copy(dAtA[i:], m.TargetActorType)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.TargetActorType)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.TargetActorID) > 0 {
		i -= len(m.TargetActorID)
		copy(dAtA[i:], m.TargetActorID)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.TargetActorID)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.PrevActorType) > 0 {
		i -= len(m.PrevActorType)
		copy(dAtA[i:], m.PrevActorType)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.PrevActorType)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.OrgActorType) > 0 {
		i -= len(m.OrgActorType)
		copy(dAtA[i:], m.OrgActorType)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.OrgActorType)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.OrgActorID) > 0 {
		i -= len(m.OrgActorID)
		copy(dAtA[i:], m.OrgActorID)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.OrgActorID)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ID) > 0 {
		i -= len(m.ID)
		copy(dAtA[i:], m.ID)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.ID)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Message) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Body) > 0 {
		i -= len(m.Body)
		copy(dAtA[i:], m.Body)
		i = encodeVarintRouter(dAtA, i, uint64(len(m.Body)))
		i--
		dAtA[i] = 0x12
	}
	if m.Header != nil {
		{
			size, err := m.Header.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintRouter(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RouteReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RouteReq) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RouteReq) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Msg != nil {
		{
			size, err := m.Msg.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintRouter(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RouteRes) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RouteRes) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RouteRes) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Msg != nil {
		{
			size, err := m.Msg.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintRouter(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}

func encodeVarintRouter(dAtA []byte, offset int, v uint64) int {
	offset -= sovRouter(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Header) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ID)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.OrgActorID)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.OrgActorType)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.PrevActorType)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.TargetActorID)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.TargetActorType)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.Event)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.Token)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	if m.Timestamp != 0 {
		n += 1 + sovRouter(uint64(m.Timestamp))
	}
	if len(m.Custom) > 0 {
		for k, v := range m.Custom {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovRouter(uint64(len(k))) + 1 + len(v) + sovRouter(uint64(len(v)))
			n += mapEntrySize + 1 + sovRouter(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *Message) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Header != nil {
		l = m.Header.Size()
		n += 1 + l + sovRouter(uint64(l))
	}
	l = len(m.Body)
	if l > 0 {
		n += 1 + l + sovRouter(uint64(l))
	}
	return n
}

func (m *RouteReq) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Msg != nil {
		l = m.Msg.Size()
		n += 1 + l + sovRouter(uint64(l))
	}
	return n
}

func (m *RouteRes) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Msg != nil {
		l = m.Msg.Size()
		n += 1 + l + sovRouter(uint64(l))
	}
	return n
}

func sovRouter(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozRouter(x uint64) (n int) {
	return sovRouter(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Header) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRouter
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Header: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Header: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrgActorID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrgActorID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrgActorType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrgActorType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PrevActorType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PrevActorType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetActorID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TargetActorID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetActorType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TargetActorType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Event", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Event = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Custom", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Custom == nil {
				m.Custom = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRouter
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRouter
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthRouter
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthRouter
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRouter
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthRouter
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthRouter
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipRouter(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthRouter
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Custom[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRouter(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRouter
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
func (m *Message) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRouter
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Header", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Header == nil {
				m.Header = &Header{}
			}
			if err := m.Header.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Body", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Body = append(m.Body[:0], dAtA[iNdEx:postIndex]...)
			if m.Body == nil {
				m.Body = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRouter(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRouter
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
func (m *RouteReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRouter
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: routeReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: routeReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Msg", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Msg == nil {
				m.Msg = &Message{}
			}
			if err := m.Msg.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRouter(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRouter
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
func (m *RouteRes) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRouter
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: routeRes: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: routeRes: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Msg", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRouter
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRouter
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Msg == nil {
				m.Msg = &Message{}
			}
			if err := m.Msg.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRouter(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRouter
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
func skipRouter(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRouter
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
					return 0, ErrIntOverflowRouter
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRouter
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
			if length < 0 {
				return 0, ErrInvalidLengthRouter
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupRouter
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthRouter
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthRouter        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRouter          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupRouter = fmt.Errorf("proto: unexpected end of group")
)