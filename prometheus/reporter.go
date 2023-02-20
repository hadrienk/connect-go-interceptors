package prometheus

import (
	"github.com/bufbuild/connect-go"
	"strconv"
	"strings"
	"time"
)

type grpcReporter struct {
	metrics *interceptor
	start   time.Time
	spec    connect.Spec
}

type reporter interface {
	monitorStart()
	monitorSend()
	monitorReceive()
	monitorDone(err error)
}

var sinceFunc = time.Since
var nowFunc = time.Now

func newReporter(spec connect.Spec, metrics *interceptor) reporter {
	return &grpcReporter{
		metrics: metrics,
		spec:    spec,
	}
}

func (p *grpcReporter) labelValues() (streamType, serviceName, methodName string) {
	serviceName, methodName = splitMethodName(p.spec.Procedure)
	streamType = streamTypeString(p.spec.StreamType)
	return
}

func (p *grpcReporter) monitorStart() {
	p.metrics.serverStartedCounter.WithLabelValues(p.labelValues()).Inc()
	if p.metrics.serverHandledHistogramEnabled {
		p.start = nowFunc()
	}
}

func (p *grpcReporter) monitorSend() {
	p.metrics.serverStreamMsgSent.WithLabelValues(p.labelValues()).Inc()
}

func (p *grpcReporter) monitorReceive() {
	p.metrics.serverStreamMsgReceived.WithLabelValues(p.labelValues()).Inc()
}

func (p *grpcReporter) monitorDone(err error) {
	streamType, serviceName, methodName := p.labelValues()
	if p.metrics.serverHandledHistogramEnabled {
		p.metrics.serverHandledHistogram.WithLabelValues(streamType, serviceName, methodName).Observe(sinceFunc(p.start).Seconds())
	}
	p.metrics.serverHandledCounter.WithLabelValues(streamType, serviceName, methodName, errorString(err)).Inc()
}

func errorString(err error) string {
	if err == nil {
		return "OK"
	}
	code := connect.CodeOf(err)
	switch code {
	case connect.CodeCanceled:
		return "Canceled"
	case connect.CodeUnknown:
		return "Unknown"
	case connect.CodeInvalidArgument:
		return "InvalidArgument"
	case connect.CodeDeadlineExceeded:
		return "DeadlineExceeded"
	case connect.CodeNotFound:
		return "NotFound"
	case connect.CodeAlreadyExists:
		return "AlreadyExists"
	case connect.CodePermissionDenied:
		return "PermissionDenied"
	case connect.CodeResourceExhausted:
		return "ResourceExhausted"
	case connect.CodeFailedPrecondition:
		return "FailedPrecondition"
	case connect.CodeAborted:
		return "Aborted"
	case connect.CodeOutOfRange:
		return "OutOfRange"
	case connect.CodeUnimplemented:
		return "Unimplemented"
	case connect.CodeInternal:
		return "Internal"
	case connect.CodeUnavailable:
		return "Unavailable"
	case connect.CodeDataLoss:
		return "DataLoss"
	case connect.CodeUnauthenticated:
		return "Unauthenticated"
	default:
		return "Code(" + strconv.FormatInt(int64(code), 10) + ")"
	}
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func streamTypeString(streamType connect.StreamType) string {
	switch {
	case streamType == connect.StreamTypeUnary:
		return "unary"
	case streamType == connect.StreamTypeClient:
		return "client_stream"
	case streamType == connect.StreamTypeServer:
		return "server_stream"
	case streamType == connect.StreamTypeBidi:
		return "bidi_stream"
	default:
		return "unknown"
	}
}
