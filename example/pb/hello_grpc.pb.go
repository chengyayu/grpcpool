// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: pb/hello.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EchoClient is the client API for Echo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EchoClient interface {
	// Say is simple request.
	Say(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error)
}

type echoClient struct {
	cc grpc.ClientConnInterface
}

func NewEchoClient(cc grpc.ClientConnInterface) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Say(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, "/pb.Echo/Say", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EchoServer is the server API for Echo service.
// All implementations must embed UnimplementedEchoServer
// for forward compatibility
type EchoServer interface {
	// Say is simple request.
	Say(context.Context, *EchoRequest) (*EchoResponse, error)
	mustEmbedUnimplementedEchoServer()
}

// UnimplementedEchoServer must be embedded to have forward compatible implementations.
type UnimplementedEchoServer struct {
}

func (UnimplementedEchoServer) Say(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Say not implemented")
}
func (UnimplementedEchoServer) mustEmbedUnimplementedEchoServer() {}

// UnsafeEchoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EchoServer will
// result in compilation errors.
type UnsafeEchoServer interface {
	mustEmbedUnimplementedEchoServer()
}

func RegisterEchoServer(s grpc.ServiceRegistrar, srv EchoServer) {
	s.RegisterService(&Echo_ServiceDesc, srv)
}

func _Echo_Say_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Say(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Echo/Say",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Say(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Echo_ServiceDesc is the grpc.ServiceDesc for Echo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Echo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Say",
			Handler:    _Echo_Say_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/hello.proto",
}
