package gerrors

import (
	"fmt"
	"io"

	. "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type withCode struct {
	cause error
	code  codes.Code
}

func (w *withCode) Error() string {
	return w.code.String() + ": " + w.cause.Error()
}
func (w *withCode) Cause() error { return w.cause }

func (w *withCode) Code() codes.Code { return w.code }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withCode) Unwrap() error { return w.cause }
func (w *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			fmt.Fprintf(s, "code: %s\n", w.code.String())
			return
		}
		fmt.Fprintf(s, "%v\n", w.Cause())
		fmt.Fprintf(s, "code: %s\n", w.code.String())
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

func NewWithCode(code codes.Code, format string, args ...any) error {
	return &withCode{
		cause: New(fmt.Sprintf(format, args...)),
		code:  code,
	}
}

func NewWithCodeAndInfo(code codes.Code, info Info, format string, args ...any) error {
	return &withInfo{
		cause: NewWithCode(code, format, args...),
		info:  info,
	}
}

func WithStackOnce(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(interface {
		StackTrace() StackTrace
	}); ok {
		return err
	}

	return WithStack(err)
}

func WithInfo(err error, info Info) error {
	if err == nil {
		return nil
	}
	return &withInfo{
		cause: err,
		info:  info,
	}
}

type Info struct {
	Domain   string
	Reason   string
	Metadata map[string]string
}

type withInfo struct {
	cause error
	info  Info
}

func (w *withInfo) Error() string {
	if w.info.Domain == "" {
		return w.info.Reason + ": " + w.cause.Error()
	}
	return fmt.Sprintf("%s.%s: %s", w.info.Domain, w.info.Reason, w.cause.Error())
}
func (w *withInfo) Cause() error { return w.cause }
func (w *withInfo) Info() Info   { return w.info }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withInfo) Unwrap() error { return w.cause }
func (w *withInfo) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			fmt.Fprintf(s, "info.domain: %s\ninfo.reason: %s\ninfo.metadata: %+v", w.info.Domain, w.info.Reason, w.info.Metadata)
			return
		}
		fmt.Fprintf(s, "%v\n", w.Cause())
		fmt.Fprintf(s, "info.domain: %s\ninfo.reason: %s\ninfo.metadata: %+v", w.info.Domain, w.info.Reason, w.info.Metadata)
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

/**
 * @title: Get GRPC Code
 * @description: 支持自定义的 error.Code，并且兼容 GRPC 原生 status.Code
 */
func Code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	var e interface{ Code() codes.Code }
	if As(err, &e) {
		return e.Code()
	}

	if se, ok := err.(interface {
		GRPCStatus() *status.Status
	}); ok {
		return se.GRPCStatus().Code()
	}
	return codes.Unknown
}
