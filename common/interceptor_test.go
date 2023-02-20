package common

import (
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestContextHeaderInterceptor_WrapStreamingClient(t *testing.T) {
	interceptor := ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
		t.Fail()
		return nil, nil
	})
	reached := false
	next := connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		reached = true
		return nil
	})
	interceptor.WrapStreamingClient(next)(nil, connect.Spec{})
	assert.True(t, reached)
}

func TestContextHeaderInterceptor_WrapStreamingHandler(t *testing.T) {
	t.Run("propagate any error", func(t *testing.T) {
		expectedError := errors.New("error")
		interceptor := ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
			return nil, expectedError
		})
		err := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			t.Fail()
			return nil
		})(context.Background(), noOpConn{})
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("passes context and header", func(t *testing.T) {
		expectedHeader := http.Header{
			"Foo": {"Bar", "Baz"},
		}
		interceptor := ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
			assert.Equal(t, expectedHeader, header)
			return context.WithValue(ctx, "test", "value"), nil
		})
		err := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			assert.Equal(t, "value", ctx.Value("test"))
			return nil
		})(context.Background(), noOpConn{reqHeader: expectedHeader})
		assert.NoError(t, err)
	})
}

func TestContextHeaderInterceptor_WrapUnary(t *testing.T) {
	t.Run("propagate any error", func(t *testing.T) {
		expectedError := errors.New("error")
		interceptor := ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
			return nil, expectedError
		})
		_, err := interceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			t.Fail()
			return nil, nil
		})(context.Background(), connect.NewRequest(&map[string]string{}))
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("passes context and header", func(t *testing.T) {
		expectedHeader := http.Header{
			"Foo": {"Bar", "Baz"},
		}

		interceptor := ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
			assert.Equal(t, expectedHeader, header)
			return context.WithValue(ctx, "test", "value"), nil
		})
		request := connect.NewRequest(&map[string]string{})
		request.Header().Add("Foo", "Bar")
		request.Header().Add("Foo", "Baz")
		_, err := interceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.Equal(t, "value", ctx.Value("test"))
			return nil, nil
		})(context.Background(), request)
		assert.NoError(t, err)
	})
}

func TestContextSpecInterceptor_WrapStreamingClient(t *testing.T) {
	interceptor := ContextSpecInterceptor(func(ctx context.Context, spec connect.Spec) (context.Context, error) {
		t.Fail()
		return nil, nil
	})
	reached := false
	next := connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		reached = true
		return nil
	})
	interceptor.WrapStreamingClient(next)(nil, connect.Spec{})
	assert.True(t, reached)
}

func TestContextSpecInterceptor_WrapStreamingHandler(t *testing.T) {
	t.Run("propagate any error", func(t *testing.T) {
		expectedError := errors.New("error")
		interceptor := ContextSpecInterceptor(func(ctx context.Context, spec connect.Spec) (context.Context, error) {
			return nil, expectedError
		})
		err := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			t.Fail()
			return nil
		})(context.Background(), noOpConn{})
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("passes context and spec", func(t *testing.T) {
		expectedSpec := connect.Spec{
			Procedure: "Foo",
		}
		interceptor := ContextSpecInterceptor(func(ctx context.Context, spec connect.Spec) (context.Context, error) {
			assert.Equal(t, expectedSpec, spec)
			return context.WithValue(ctx, "test", "value"), nil
		})
		err := interceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			assert.Equal(t, "value", ctx.Value("test"))
			return nil
		})(context.Background(), noOpConn{spec: expectedSpec})
		assert.NoError(t, err)
	})
}

func TestContextSpecInterceptor_WrapUnary(t *testing.T) {
	t.Run("propagate any error", func(t *testing.T) {
		expectedError := errors.New("error")
		interceptor := ContextSpecInterceptor(func(ctx context.Context, spec connect.Spec) (context.Context, error) {
			return nil, expectedError
		})
		_, err := interceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			t.Fail()
			return nil, nil
		})(context.Background(), connect.NewRequest(&map[string]string{}))
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("passes context and header", func(t *testing.T) {
		request := connect.NewRequest(&map[string]string{})
		interceptor := ContextSpecInterceptor(func(ctx context.Context, spec connect.Spec) (context.Context, error) {
			assert.Equal(t, request.Spec(), spec)
			return context.WithValue(ctx, "test", "value"), nil
		})
		_, err := interceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.Equal(t, "value", ctx.Value("test"))
			return nil, nil
		})(context.Background(), request)
		assert.NoError(t, err)
	})
}

type noOpConn struct {
	receive    func(msg any) error
	send       func(msg any) error
	reqHeader  http.Header
	respHeader http.Header
	spec       connect.Spec
}

func (n noOpConn) Spec() connect.Spec {
	return n.spec
}

func (n noOpConn) Peer() connect.Peer {
	return connect.Peer{}
}

func (n noOpConn) Receive(msg any) error {
	if n.receive != nil {
		return n.receive(msg)
	}
	return nil
}

func (n noOpConn) RequestHeader() http.Header {
	return n.reqHeader
}

func (n noOpConn) Send(msg any) error {
	if n.receive != nil {
		return n.send(msg)
	}
	return nil
}

func (n noOpConn) ResponseHeader() http.Header {
	return n.respHeader

}

func (n noOpConn) ResponseTrailer() http.Header {
	return nil
}
