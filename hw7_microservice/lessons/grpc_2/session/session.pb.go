// Code generated by protoc-gen-go. DO NOT EDIT.
// source: session.proto

package session

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type SessionID struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	//XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache int32 `json:"-"`
}

func (m *SessionID) Reset()         { *m = SessionID{} }
func (m *SessionID) String() string { return proto.CompactTextString(m) }
func (*SessionID) ProtoMessage()    {}
func (*SessionID) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a6be1b361fa6f14, []int{0}
}

func (m *SessionID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionID.Unmarshal(m, b)
}
func (m *SessionID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionID.Marshal(b, m, deterministic)
}
func (m *SessionID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionID.Merge(m, src)
}
func (m *SessionID) XXX_Size() int {
	return xxx_messageInfo_SessionID.Size(m)
}
func (m *SessionID) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionID.DiscardUnknown(m)
}

var xxx_messageInfo_SessionID proto.InternalMessageInfo

func (m *SessionID) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type Session struct {
	Login                string   `protobuf:"bytes,1,opt,name=login,proto3" json:"login,omitempty"`
	Useragent            string   `protobuf:"bytes,2,opt,name=useragent,proto3" json:"useragent,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Session) Reset()         { *m = Session{} }
func (m *Session) String() string { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()    {}
func (*Session) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a6be1b361fa6f14, []int{1}
}

func (m *Session) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Session.Unmarshal(m, b)
}
func (m *Session) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Session.Marshal(b, m, deterministic)
}
func (m *Session) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Session.Merge(m, src)
}
func (m *Session) XXX_Size() int {
	return xxx_messageInfo_Session.Size(m)
}
func (m *Session) XXX_DiscardUnknown() {
	xxx_messageInfo_Session.DiscardUnknown(m)
}

var xxx_messageInfo_Session proto.InternalMessageInfo

func (m *Session) GetLogin() string {
	if m != nil {
		return m.Login
	}
	return ""
}

func (m *Session) GetUseragent() string {
	if m != nil {
		return m.Useragent
	}
	return ""
}

type Nothing struct {
	Dummy                bool     `protobuf:"varint,1,opt,name=dummy,proto3" json:"dummy,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nothing) Reset()         { *m = Nothing{} }
func (m *Nothing) String() string { return proto.CompactTextString(m) }
func (*Nothing) ProtoMessage()    {}
func (*Nothing) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a6be1b361fa6f14, []int{2}
}

func (m *Nothing) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nothing.Unmarshal(m, b)
}
func (m *Nothing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nothing.Marshal(b, m, deterministic)
}
func (m *Nothing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nothing.Merge(m, src)
}
func (m *Nothing) XXX_Size() int {
	return xxx_messageInfo_Nothing.Size(m)
}
func (m *Nothing) XXX_DiscardUnknown() {
	xxx_messageInfo_Nothing.DiscardUnknown(m)
}

var xxx_messageInfo_Nothing proto.InternalMessageInfo

func (m *Nothing) GetDummy() bool {
	if m != nil {
		return m.Dummy
	}
	return false
}

func init() {
	proto.RegisterType((*SessionID)(nil), "session.SessionID")
	proto.RegisterType((*Session)(nil), "session.Session")
	proto.RegisterType((*Nothing)(nil), "session.Nothing")
}

func init() { proto.RegisterFile("session.proto", fileDescriptor_3a6be1b361fa6f14) }

var fileDescriptor_3a6be1b361fa6f14 = []byte{
	// 205 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4e, 0x2d, 0x2e,
	0xce, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0x95, 0xa4, 0xb9,
	0x38, 0x83, 0x21, 0x4c, 0x4f, 0x17, 0x21, 0x3e, 0x2e, 0x26, 0x4f, 0x17, 0x09, 0x46, 0x05, 0x46,
	0x0d, 0xce, 0x20, 0x26, 0x4f, 0x17, 0x25, 0x5b, 0x2e, 0x76, 0xa8, 0xa4, 0x90, 0x08, 0x17, 0x6b,
	0x4e, 0x7e, 0x7a, 0x66, 0x1e, 0x54, 0x16, 0xc2, 0x11, 0x92, 0xe1, 0xe2, 0x2c, 0x2d, 0x4e, 0x2d,
	0x4a, 0x4c, 0x4f, 0xcd, 0x2b, 0x91, 0x60, 0x02, 0xcb, 0x20, 0x04, 0x94, 0xe4, 0xb9, 0xd8, 0xfd,
	0xf2, 0x4b, 0x32, 0x32, 0xf3, 0xd2, 0x41, 0xda, 0x53, 0x4a, 0x73, 0x73, 0x2b, 0xc1, 0xda, 0x39,
	0x82, 0x20, 0x1c, 0xa3, 0x45, 0x8c, 0x5c, 0xdc, 0x8e, 0xa5, 0x25, 0x19, 0xce, 0x19, 0xa9, 0xc9,
	0xd9, 0xa9, 0x45, 0x42, 0x06, 0x5c, 0x6c, 0xce, 0x45, 0xa9, 0x89, 0x25, 0xa9, 0x42, 0x02, 0x7a,
	0x30, 0xf7, 0x42, 0x1d, 0x20, 0x25, 0x84, 0x2e, 0xe2, 0xe9, 0xa2, 0xc4, 0x20, 0xa4, 0xcf, 0xc5,
	0x0a, 0xd6, 0x2c, 0x84, 0x45, 0x5a, 0x0a, 0xc3, 0x10, 0x25, 0x06, 0x90, 0x15, 0x2e, 0xa9, 0x39,
	0xa9, 0x25, 0xa9, 0x04, 0x74, 0x40, 0x1d, 0xae, 0xc4, 0x90, 0xc4, 0x06, 0x0e, 0x31, 0x63, 0x40,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x7f, 0x14, 0x16, 0x5c, 0x42, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AuthCheckerClient is the client API for AuthChecker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AuthCheckerClient interface {
	Create(ctx context.Context, in *Session, opts ...grpc.CallOption) (*SessionID, error)
	Check(ctx context.Context, in *SessionID, opts ...grpc.CallOption) (*Session, error)
	Delete(ctx context.Context, in *SessionID, opts ...grpc.CallOption) (*Nothing, error)
}

type authCheckerClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthCheckerClient(cc grpc.ClientConnInterface) AuthCheckerClient {
	return &authCheckerClient{cc}
}

func (c *authCheckerClient) Create(ctx context.Context, in *Session, opts ...grpc.CallOption) (*SessionID, error) {
	out := new(SessionID)
	err := c.cc.Invoke(ctx, "/session.AuthChecker/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authCheckerClient) Check(ctx context.Context, in *SessionID, opts ...grpc.CallOption) (*Session, error) {
	out := new(Session)
	err := c.cc.Invoke(ctx, "/session.AuthChecker/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authCheckerClient) Delete(ctx context.Context, in *SessionID, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/session.AuthChecker/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthCheckerServer is the server API for AuthChecker service.
type AuthCheckerServer interface {
	Create(context.Context, *Session) (*SessionID, error)
	Check(context.Context, *SessionID) (*Session, error)
	Delete(context.Context, *SessionID) (*Nothing, error)
}

// UnimplementedAuthCheckerServer can be embedded to have forward compatible implementations.
type UnimplementedAuthCheckerServer struct {
}

func (*UnimplementedAuthCheckerServer) Create(ctx context.Context, req *Session) (*SessionID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (*UnimplementedAuthCheckerServer) Check(ctx context.Context, req *SessionID) (*Session, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}
func (*UnimplementedAuthCheckerServer) Delete(ctx context.Context, req *SessionID) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func RegisterAuthCheckerServer(s *grpc.Server, srv AuthCheckerServer) {
	s.RegisterService(&_AuthChecker_serviceDesc, srv)
}

func _AuthChecker_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Session)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthCheckerServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/session.AuthChecker/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthCheckerServer).Create(ctx, req.(*Session))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthChecker_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthCheckerServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/session.AuthChecker/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthCheckerServer).Check(ctx, req.(*SessionID))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthChecker_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthCheckerServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/session.AuthChecker/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthCheckerServer).Delete(ctx, req.(*SessionID))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthChecker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "session.AuthChecker",
	HandlerType: (*AuthCheckerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _AuthChecker_Create_Handler,
		},
		{
			MethodName: "Check",
			Handler:    _AuthChecker_Check_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _AuthChecker_Delete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "session.proto",
}
