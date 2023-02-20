package prometheus

import (
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_errorString(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"handles CodeCanceled", args{connect.NewError(connect.CodeCanceled, io.EOF)}, "Canceled"},
		{"handles CodeUnknown", args{connect.NewError(connect.CodeUnknown, io.EOF)}, "Unknown"},
		{"handles CodeInvalidArgument", args{connect.NewError(connect.CodeInvalidArgument, io.EOF)}, "InvalidArgument"},
		{"handles CodeDeadlineExceeded", args{connect.NewError(connect.CodeDeadlineExceeded, io.EOF)}, "DeadlineExceeded"},
		{"handles CodeNotFound", args{connect.NewError(connect.CodeNotFound, io.EOF)}, "NotFound"},
		{"handles CodeAlreadyExists", args{connect.NewError(connect.CodeAlreadyExists, io.EOF)}, "AlreadyExists"},
		{"handles CodePermissionDenied", args{connect.NewError(connect.CodePermissionDenied, io.EOF)}, "PermissionDenied"},
		{"handles CodeResourceExhausted", args{connect.NewError(connect.CodeResourceExhausted, io.EOF)}, "ResourceExhausted"},
		{"handles CodeFailedPrecondition", args{connect.NewError(connect.CodeFailedPrecondition, io.EOF)}, "FailedPrecondition"},
		{"handles CodeAborted", args{connect.NewError(connect.CodeAborted, io.EOF)}, "Aborted"},
		{"handles CodeOutOfRange", args{connect.NewError(connect.CodeOutOfRange, io.EOF)}, "OutOfRange"},
		{"handles CodeUnimplemented", args{connect.NewError(connect.CodeUnimplemented, io.EOF)}, "Unimplemented"},
		{"handles CodeInternal", args{connect.NewError(connect.CodeInternal, io.EOF)}, "Internal"},
		{"handles CodeUnavailable", args{connect.NewError(connect.CodeUnavailable, io.EOF)}, "Unavailable"},
		{"handles CodeDataLoss", args{connect.NewError(connect.CodeDataLoss, io.EOF)}, "DataLoss"},
		{"handles CodeUnauthenticated", args{connect.NewError(connect.CodeUnauthenticated, io.EOF)}, "Unauthenticated"},
		{"handles other codes", args{connect.NewError(1234, io.EOF)}, "Code(1234)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, errorString(tt.args.err), "errorString(%v)", tt.args.err)
		})
	}
}

func Test_streamTypeString(t *testing.T) {
	type args struct {
		streamType connect.StreamType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"handles connect.StreamTypeUnary", args{connect.StreamTypeUnary}, "unary"},
		{"handles connect.StreamTypeClient", args{connect.StreamTypeClient}, "client_stream"},
		{"handles connect.StreamTypeServer", args{connect.StreamTypeServer}, "server_stream"},
		{"handles connect.StreamTypeBidi", args{connect.StreamTypeBidi}, "bidi_stream"},
		{"handles other type", args{0xFF}, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, streamTypeString(tt.args.streamType), "streamTypeString(%v)", tt.args.streamType)
		})
	}
}

func Test_splitMethodName(t *testing.T) {
	type args struct {
		fullMethodName string
	}
	tests := []struct {
		name        string
		args        args
		wantPackage string
		wantMethod  string
	}{
		{"correct package & method", args{"foo.bar/Baz"}, "foo.bar", "Baz"},
		{"empty method name", args{"foo.bar/"}, "foo.bar", ""},
		{"empty package", args{"/Baz"}, "unknown", "unknown"},
		{"only package", args{"foo.bar"}, "unknown", "unknown"},
		{"only method", args{"Baz"}, "unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitMethodName(tt.args.fullMethodName)
			assert.Equalf(t, tt.wantPackage, got, "splitMethodName(%v)", tt.args.fullMethodName)
			assert.Equalf(t, tt.wantMethod, got1, "splitMethodName(%v)", tt.args.fullMethodName)
		})
	}
}
