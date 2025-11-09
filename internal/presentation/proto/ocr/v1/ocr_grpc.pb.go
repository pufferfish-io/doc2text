package ocrv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	OcrService_Process_FullMethodName = "/ocr.v1.OcrService/Process"
)

type OcrServiceClient interface {
	Process(ctx context.Context, in *ParseRequest, opts ...grpc.CallOption) (*ParseResponse, error)
}

type ocrServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOcrServiceClient(cc grpc.ClientConnInterface) OcrServiceClient {
	return &ocrServiceClient{cc}
}

func (c *ocrServiceClient) Process(ctx context.Context, in *ParseRequest, opts ...grpc.CallOption) (*ParseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ParseResponse)
	err := c.cc.Invoke(ctx, OcrService_Process_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type OcrServiceServer interface {
	Process(context.Context, *ParseRequest) (*ParseResponse, error)
	mustEmbedUnimplementedOcrServiceServer()
}

type UnimplementedOcrServiceServer struct{}

func (UnimplementedOcrServiceServer) Process(context.Context, *ParseRequest) (*ParseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Process not implemented")
}
func (UnimplementedOcrServiceServer) mustEmbedUnimplementedOcrServiceServer() {}
func (UnimplementedOcrServiceServer) testEmbeddedByValue()                    {}

type UnsafeOcrServiceServer interface {
	mustEmbedUnimplementedOcrServiceServer()
}

func RegisterOcrServiceServer(s grpc.ServiceRegistrar, srv OcrServiceServer) {

	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OcrService_ServiceDesc, srv)
}

func _OcrService_Process_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OcrServiceServer).Process(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OcrService_Process_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OcrServiceServer).Process(ctx, req.(*ParseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var OcrService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ocr.v1.OcrService",
	HandlerType: (*OcrServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Process",
			Handler:    _OcrService_Process_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/presentation/proto/ocr/v1/ocr.proto",
}
