// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: proto/source_auth/source_auth.proto

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

const (
	SourceAuthService_PostSourceAuth_FullMethodName = "/proto.source_auth.SourceAuthService/PostSourceAuth"
)

// SourceAuthServiceClient is the client API for SourceAuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SourceAuthServiceClient interface {
	PostSourceAuth(ctx context.Context, in *SourceAuthBodyRequest, opts ...grpc.CallOption) (*SourceAuthResponse, error)
}

type sourceAuthServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSourceAuthServiceClient(cc grpc.ClientConnInterface) SourceAuthServiceClient {
	return &sourceAuthServiceClient{cc}
}

func (c *sourceAuthServiceClient) PostSourceAuth(ctx context.Context, in *SourceAuthBodyRequest, opts ...grpc.CallOption) (*SourceAuthResponse, error) {
	out := new(SourceAuthResponse)
	err := c.cc.Invoke(ctx, SourceAuthService_PostSourceAuth_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SourceAuthServiceServer is the server API for SourceAuthService service.
// All implementations should embed UnimplementedSourceAuthServiceServer
// for forward compatibility
type SourceAuthServiceServer interface {
	PostSourceAuth(context.Context, *SourceAuthBodyRequest) (*SourceAuthResponse, error)
}

// UnimplementedSourceAuthServiceServer should be embedded to have forward compatible implementations.
type UnimplementedSourceAuthServiceServer struct {
}

func (UnimplementedSourceAuthServiceServer) PostSourceAuth(context.Context, *SourceAuthBodyRequest) (*SourceAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostSourceAuth not implemented")
}

// UnsafeSourceAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SourceAuthServiceServer will
// result in compilation errors.
type UnsafeSourceAuthServiceServer interface {
	mustEmbedUnimplementedSourceAuthServiceServer()
}

func RegisterSourceAuthServiceServer(s grpc.ServiceRegistrar, srv SourceAuthServiceServer) {
	s.RegisterService(&SourceAuthService_ServiceDesc, srv)
}

func _SourceAuthService_PostSourceAuth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SourceAuthBodyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SourceAuthServiceServer).PostSourceAuth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SourceAuthService_PostSourceAuth_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SourceAuthServiceServer).PostSourceAuth(ctx, req.(*SourceAuthBodyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SourceAuthService_ServiceDesc is the grpc.ServiceDesc for SourceAuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SourceAuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.source_auth.SourceAuthService",
	HandlerType: (*SourceAuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostSourceAuth",
			Handler:    _SourceAuthService_PostSourceAuth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/source_auth/source_auth.proto",
}