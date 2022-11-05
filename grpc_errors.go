package gerrors

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		return WithStackOnce(err)
	}
}

type options struct {
	beforeReturn func(context.Context, *status.Status) *status.Status
}

type Option func(*options)

func WithBeforeReturn(beforeReturn func(context.Context, *status.Status) *status.Status) Option {
	return func(o *options) {
		o.beforeReturn = beforeReturn
	}
}

func defaultBeforeReturn(ctx context.Context, s *status.Status) *status.Status {
	ps := s.Proto()
	switch s.Code() {
	case codes.Internal, codes.Unknown:
		ps.Message = s.Code().String()
	}

	return status.FromProto(ps)
}

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := &options{
		beforeReturn: defaultBeforeReturn,
	}
	for _, fn := range opts {
		fn(o)
	}
	return func(ctx context.Context, req interface{}, serverInfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err == nil {
			return
		}

		if ss, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			return resp, o.beforeReturn(ctx, ss.GRPCStatus()).Err()
		}

		code := Code(err)
		if code == codes.OK {
			return nil, status.New(code, err.Error()).Err()
		}

		// ErrorInfo
		s := status.New(code, err.Error())

		var wi interface{ Info() Info }
		var info Info
		if errors.As(err, &wi) {
			info = wi.Info()
		}

		if info.Reason != "" {
			s, _ = s.WithDetails(&errdetails.ErrorInfo{
				Reason:   info.Reason,
				Domain:   info.Domain,
				Metadata: info.Metadata,
			})
		}

		s = o.beforeReturn(ctx, s)
		return nil, s.Err()
	}
}
