package prometheus

import (
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

type msg struct {
}

func Test_interceptor_WrapUnary(t *testing.T) {
	prometheusInterceptor := NewPrometheusInterceptor(WithConstLabels(prom.Labels{
		"foo": "bar",
	}))

	expectedRequest := connect.NewRequest(&msg{})
	expectedResponse := connect.NewResponse(&msg{})
	expectedError := errors.New("error")

	t.Run("collect metrics", func(t *testing.T) {
		response, err := prometheusInterceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.Same(t, expectedRequest, request)
			return expectedResponse, nil
		})(context.Background(), expectedRequest)
		assert.Same(t, expectedResponse, response)
		assert.NoError(t, err)

		assert.NoError(t, testutil.CollectAndCompare(prometheusInterceptor, strings.NewReader(`
			# HELP grpc_server_handled_total Total number of RPCs completed on the server, regardless of success or failure.
			# TYPE grpc_server_handled_total counter
			grpc_server_handled_total{foo="bar",grpc_code="OK",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
			# HELP grpc_server_msg_received_total Total number of RPC stream messages received on the server.
			# TYPE grpc_server_msg_received_total counter
			grpc_server_msg_received_total{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
			# HELP grpc_server_msg_sent_total Total number of gRPC stream messages sent by the server.
			# TYPE grpc_server_msg_sent_total counter
			grpc_server_msg_sent_total{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
			# HELP grpc_server_started_total Total number of RPCs started on the server.
			# TYPE grpc_server_started_total counter
			grpc_server_started_total{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
		`),
			"grpc_server_started_total",
			"grpc_server_handled_total",
			"grpc_server_msg_received_total",
			"grpc_server_msg_sent_total",
			"grpc_server_handling_seconds",
		))
	})

	t.Run("returns same error", func(t *testing.T) {
		_, err := prometheusInterceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.Same(t, expectedRequest, request)
			return nil, expectedError
		})(context.Background(), expectedRequest)
		assert.Same(t, expectedError, err)

	})

	t.Run("supports histograms", func(t *testing.T) {
		prometheusInterceptor := NewPrometheusInterceptor()
		prometheusInterceptor.EnableHandlingTimeHistogram(WithHistogramConstLabels(prom.Labels{
			"foo": "bar",
		}))

		defer func() {
			nowFunc = time.Now
			sinceFunc = time.Since
		}()

		nowFunc = func() time.Time {
			return time.Time{}
		}

		d := 16 * time.Second
		sinceFunc = func(_ time.Time) time.Duration {
			d /= 2
			return d
		}

		for i := 0; i < 32; i++ {
			response, err := prometheusInterceptor.WrapUnary(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
				assert.Same(t, expectedRequest, request)
				return expectedResponse, nil
			})(context.Background(), expectedRequest)
			assert.Same(t, expectedResponse, response)
			assert.NoError(t, err)
		}

		assert.NoError(t, testutil.CollectAndCompare(prometheusInterceptor, strings.NewReader(`
			# HELP grpc_server_handling_seconds Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.
			# TYPE grpc_server_handling_seconds histogram
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.005"} 21
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.01"} 22
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.025"} 23
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.05"} 24
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.1"} 25
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.25"} 27
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.5"} 28
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="1"} 29
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="2.5"} 30
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="5"} 31
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="10"} 32
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="+Inf"} 32
			grpc_server_handling_seconds_sum{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 15.999999986
			grpc_server_handling_seconds_count{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 32
		`), "grpc_server_handling_seconds"))

	})

}

func Test_interceptor_WrapStreamingClient(t *testing.T) {
	prometheusInterceptor := NewPrometheusInterceptor()
	var called bool
	next := connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		called = true
		return nil
	})
	prometheusInterceptor.WrapStreamingClient(next)(nil, connect.Spec{})
	assert.True(t, called)
}

func Test_interceptor_WrapStreamingHandler(t *testing.T) {
	prometheusInterceptor := NewPrometheusInterceptor()

	expectedRequest := connect.NewRequest(&msg{})
	expectedResponse := connect.NewResponse(&msg{})
	expectedError := errors.New("error")

	conn := noOpConn{
		send: func(msg any) error {
			assert.Same(t, msg, expectedResponse)
			return nil
		},
		receive: func(msg any) error {
			assert.Same(t, msg, expectedRequest)
			return nil
		},
	}
	t.Run("collect metrics", func(t *testing.T) {

		var wrapped connect.StreamingHandlerConn
		err := prometheusInterceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			wrapped = conn
			return nil
		})(context.Background(), conn)
		assert.NoError(t, err)

		err = wrapped.Receive(expectedRequest)
		assert.NoError(t, err)

		err = wrapped.Receive(expectedRequest)
		assert.NoError(t, err)

		err = wrapped.Send(expectedResponse)
		assert.NoError(t, err)

		err = wrapped.Send(expectedResponse)
		assert.NoError(t, err)

		err = wrapped.Send(expectedResponse)
		assert.NoError(t, err)

		assert.NoError(t, testutil.CollectAndCompare(prometheusInterceptor, strings.NewReader(`
			# HELP grpc_server_handled_total Total number of RPCs completed on the server, regardless of success or failure.
			# TYPE grpc_server_handled_total counter
			grpc_server_handled_total{grpc_code="OK",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
			# HELP grpc_server_msg_received_total Total number of RPC stream messages received on the server.
			# TYPE grpc_server_msg_received_total counter
			grpc_server_msg_received_total{grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 2
			# HELP grpc_server_msg_sent_total Total number of gRPC stream messages sent by the server.
			# TYPE grpc_server_msg_sent_total counter
			grpc_server_msg_sent_total{grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 3
			# HELP grpc_server_started_total Total number of RPCs started on the server.
			# TYPE grpc_server_started_total counter
			grpc_server_started_total{grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 1
			`),
			"grpc_server_started_total",
			"grpc_server_handled_total",
			"grpc_server_msg_received_total",
			"grpc_server_msg_sent_total",
			"grpc_server_handling_seconds",
		))
	})

	t.Run("returns same error", func(t *testing.T) {
		err := prometheusInterceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			return expectedError
		})(context.Background(), conn)
		assert.Same(t, expectedError, err)
	})

	t.Run("supports histograms", func(t *testing.T) {
		prometheusInterceptor := NewPrometheusInterceptor()
		prometheusInterceptor.EnableHandlingTimeHistogram(WithHistogramConstLabels(prom.Labels{
			"foo": "bar",
		}))

		defer func() {
			nowFunc = time.Now
			sinceFunc = time.Since
		}()

		nowFunc = func() time.Time {
			return time.Time{}
		}

		d := 16 * time.Second
		sinceFunc = func(_ time.Time) time.Duration {
			d /= 2
			return d
		}

		for i := 0; i < 32; i++ {
			var wrapped connect.StreamingHandlerConn
			err := prometheusInterceptor.WrapStreamingHandler(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
				wrapped = conn
				return nil
			})(context.Background(), conn)
			assert.NoError(t, err)
			_ = wrapped.Receive(expectedRequest)
			_ = wrapped.Send(expectedResponse)

		}

		assert.NoError(t, testutil.CollectAndCompare(prometheusInterceptor, strings.NewReader(`
			# HELP grpc_server_handling_seconds Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.
			# TYPE grpc_server_handling_seconds histogram
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.005"} 21
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.01"} 22
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.025"} 23
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.05"} 24
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.1"} 25
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.25"} 27
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="0.5"} 28
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="1"} 29
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="2.5"} 30
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="5"} 31
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="10"} 32
			grpc_server_handling_seconds_bucket{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary",le="+Inf"} 32
			grpc_server_handling_seconds_sum{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 15.999999986
			grpc_server_handling_seconds_count{foo="bar",grpc_method="unknown",grpc_service="unknown",grpc_type="unary"} 32
		`), "grpc_server_handling_seconds"))

	})

}

type noOpConn struct {
	receive func(msg any) error
	send    func(msg any) error
}

func (n noOpConn) Spec() connect.Spec {
	return connect.Spec{}
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
	return http.Header{}
}

func (n noOpConn) Send(msg any) error {
	if n.receive != nil {
		return n.send(msg)
	}
	return nil
}

func (n noOpConn) ResponseHeader() http.Header {
	return http.Header{}

}

func (n noOpConn) ResponseTrailer() http.Header {
	return http.Header{}
}
