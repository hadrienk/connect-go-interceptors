package prometheus

import (
	"context"
	"github.com/bufbuild/connect-go"
	prom "github.com/prometheus/client_golang/prometheus"
)

type interceptor struct {
	serverStartedCounter          *prom.CounterVec
	serverHandledCounter          *prom.CounterVec
	serverStreamMsgReceived       *prom.CounterVec
	serverStreamMsgSent           *prom.CounterVec
	serverHandledHistogramEnabled bool
	serverHandledHistogramOpts    prom.HistogramOpts
	serverHandledHistogram        *prom.HistogramVec
}

func (p *interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		r := newReporter(request.Spec(), p)
		r.monitorStart()
		r.monitorReceive()
		response, err := next(ctx, request)
		r.monitorDone(err)
		if err == nil {
			r.monitorSend()
		}
		return response, err
	}
}

func (p *interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (p *interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		r := newReporter(conn.Spec(), p)
		r.monitorStart()
		err := next(ctx, &monitoringHandler{
			StreamingHandlerConn: conn,
			reporter:             r,
		})
		r.monitorDone(err)
		return err
	}
}
