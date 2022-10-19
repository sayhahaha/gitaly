// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: middleware/limithandler/testdata/test.proto

package testdata

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

// TestClient is the client API for Test service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestClient interface {
	Unary(ctx context.Context, in *UnaryRequest, opts ...grpc.CallOption) (*UnaryResponse, error)
	StreamInput(ctx context.Context, opts ...grpc.CallOption) (Test_StreamInputClient, error)
	StreamOutput(ctx context.Context, in *StreamOutputRequest, opts ...grpc.CallOption) (Test_StreamOutputClient, error)
	Bidirectional(ctx context.Context, opts ...grpc.CallOption) (Test_BidirectionalClient, error)
}

type testClient struct {
	cc grpc.ClientConnInterface
}

func NewTestClient(cc grpc.ClientConnInterface) TestClient {
	return &testClient{cc}
}

func (c *testClient) Unary(ctx context.Context, in *UnaryRequest, opts ...grpc.CallOption) (*UnaryResponse, error) {
	out := new(UnaryResponse)
	err := c.cc.Invoke(ctx, "/test.limithandler.Test/Unary", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testClient) StreamInput(ctx context.Context, opts ...grpc.CallOption) (Test_StreamInputClient, error) {
	stream, err := c.cc.NewStream(ctx, &Test_ServiceDesc.Streams[0], "/test.limithandler.Test/StreamInput", opts...)
	if err != nil {
		return nil, err
	}
	x := &testStreamInputClient{stream}
	return x, nil
}

type Test_StreamInputClient interface {
	Send(*StreamInputRequest) error
	CloseAndRecv() (*StreamInputResponse, error)
	grpc.ClientStream
}

type testStreamInputClient struct {
	grpc.ClientStream
}

func (x *testStreamInputClient) Send(m *StreamInputRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *testStreamInputClient) CloseAndRecv() (*StreamInputResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(StreamInputResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *testClient) StreamOutput(ctx context.Context, in *StreamOutputRequest, opts ...grpc.CallOption) (Test_StreamOutputClient, error) {
	stream, err := c.cc.NewStream(ctx, &Test_ServiceDesc.Streams[1], "/test.limithandler.Test/StreamOutput", opts...)
	if err != nil {
		return nil, err
	}
	x := &testStreamOutputClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Test_StreamOutputClient interface {
	Recv() (*StreamOutputResponse, error)
	grpc.ClientStream
}

type testStreamOutputClient struct {
	grpc.ClientStream
}

func (x *testStreamOutputClient) Recv() (*StreamOutputResponse, error) {
	m := new(StreamOutputResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *testClient) Bidirectional(ctx context.Context, opts ...grpc.CallOption) (Test_BidirectionalClient, error) {
	stream, err := c.cc.NewStream(ctx, &Test_ServiceDesc.Streams[2], "/test.limithandler.Test/Bidirectional", opts...)
	if err != nil {
		return nil, err
	}
	x := &testBidirectionalClient{stream}
	return x, nil
}

type Test_BidirectionalClient interface {
	Send(*BidirectionalRequest) error
	Recv() (*BidirectionalResponse, error)
	grpc.ClientStream
}

type testBidirectionalClient struct {
	grpc.ClientStream
}

func (x *testBidirectionalClient) Send(m *BidirectionalRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *testBidirectionalClient) Recv() (*BidirectionalResponse, error) {
	m := new(BidirectionalResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TestServer is the server API for Test service.
// All implementations must embed UnimplementedTestServer
// for forward compatibility
type TestServer interface {
	Unary(context.Context, *UnaryRequest) (*UnaryResponse, error)
	StreamInput(Test_StreamInputServer) error
	StreamOutput(*StreamOutputRequest, Test_StreamOutputServer) error
	Bidirectional(Test_BidirectionalServer) error
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestServer must be embedded to have forward compatible implementations.
type UnimplementedTestServer struct {
}

func (UnimplementedTestServer) Unary(context.Context, *UnaryRequest) (*UnaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unary not implemented")
}
func (UnimplementedTestServer) StreamInput(Test_StreamInputServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamInput not implemented")
}
func (UnimplementedTestServer) StreamOutput(*StreamOutputRequest, Test_StreamOutputServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamOutput not implemented")
}
func (UnimplementedTestServer) Bidirectional(Test_BidirectionalServer) error {
	return status.Errorf(codes.Unimplemented, "method Bidirectional not implemented")
}
func (UnimplementedTestServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestServer will
// result in compilation errors.
type UnsafeTestServer interface {
	mustEmbedUnimplementedTestServer()
}

func RegisterTestServer(s grpc.ServiceRegistrar, srv TestServer) {
	s.RegisterService(&Test_ServiceDesc, srv)
}

func _Test_Unary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnaryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).Unary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/test.limithandler.Test/Unary",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).Unary(ctx, req.(*UnaryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Test_StreamInput_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServer).StreamInput(&testStreamInputServer{stream})
}

type Test_StreamInputServer interface {
	SendAndClose(*StreamInputResponse) error
	Recv() (*StreamInputRequest, error)
	grpc.ServerStream
}

type testStreamInputServer struct {
	grpc.ServerStream
}

func (x *testStreamInputServer) SendAndClose(m *StreamInputResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *testStreamInputServer) Recv() (*StreamInputRequest, error) {
	m := new(StreamInputRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Test_StreamOutput_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamOutputRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TestServer).StreamOutput(m, &testStreamOutputServer{stream})
}

type Test_StreamOutputServer interface {
	Send(*StreamOutputResponse) error
	grpc.ServerStream
}

type testStreamOutputServer struct {
	grpc.ServerStream
}

func (x *testStreamOutputServer) Send(m *StreamOutputResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Test_Bidirectional_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TestServer).Bidirectional(&testBidirectionalServer{stream})
}

type Test_BidirectionalServer interface {
	Send(*BidirectionalResponse) error
	Recv() (*BidirectionalRequest, error)
	grpc.ServerStream
}

type testBidirectionalServer struct {
	grpc.ServerStream
}

func (x *testBidirectionalServer) Send(m *BidirectionalResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *testBidirectionalServer) Recv() (*BidirectionalRequest, error) {
	m := new(BidirectionalRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Test_ServiceDesc is the grpc.ServiceDesc for Test service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Test_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.limithandler.Test",
	HandlerType: (*TestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Unary",
			Handler:    _Test_Unary_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamInput",
			Handler:       _Test_StreamInput_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamOutput",
			Handler:       _Test_StreamOutput_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Bidirectional",
			Handler:       _Test_Bidirectional_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "middleware/limithandler/testdata/test.proto",
}
