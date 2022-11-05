package gerrors

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewWithCode(t *testing.T) {
	a := assert.New(t)
	err := NewWithCode(codes.InvalidArgument, "code invalid")
	a.Equal(codes.InvalidArgument, Code(err))
}

func TestNewWithCodeAndInfo(t *testing.T) {
	a := assert.New(t)
	err := NewWithCodeAndInfo(codes.FailedPrecondition, Info{
		Domain: "github.com",
	}, "failed precondition")
	a.Equal(codes.FailedPrecondition, Code(err))
}

func TestWithInfo(t *testing.T) {
	a := assert.New(t)
	err := NewWithCode(codes.InvalidArgument, "code invalid")
	err = WithInfo(err, Info{
		Domain: "test.localhost",
		Reason: "TEST_VALID_INFO",
		Metadata: map[string]string{
			"a": "b",
		},
	})

	s := status.New(Code(err), err.Error())

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

	for _, detail := range s.Details() {
		if ei, ok := detail.(*errdetails.ErrorInfo); ok {
			a.Equal("test.localhost", ei.Domain)
		}
	}
}
