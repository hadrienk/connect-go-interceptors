package common

import (
	"context"
	"github.com/bufbuild/connect-go"
	"net/http"
)

// ContextSpecInterceptor is a server interceptor that fails all call with the error it returns.
type ContextSpecInterceptor func(ctx context.Context, spec connect.Spec) (context.Context, error)

func (csi ContextSpecInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := csi(ctx, request.Spec())
		if err != nil {
			return nil, err
		}
		return next(ctx, request)
	}
}

func (csi ContextSpecInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (csi ContextSpecInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := csi(ctx, conn.Spec())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

// ContextHeaderInterceptor is a server interceptor that fails all call with the error it returns.
type ContextHeaderInterceptor func(ctx context.Context, header http.Header) (context.Context, error)

func (chi ContextHeaderInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := chi(ctx, request.Header())
		if err != nil {
			return nil, err
		}
		return next(ctx, request)
	}
}

func (chi ContextHeaderInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (chi ContextHeaderInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := chi(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}
