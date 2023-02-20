package prometheus

import (
	"github.com/bufbuild/connect-go"
	prom "github.com/prometheus/client_golang/prometheus"
)

type PrometheusInterceptor interface {
	prom.Collector
	connect.Interceptor
	EnableHandlingTimeHistogram(opts ...HistogramOption)
}

// NewPrometheusInterceptor returns a PrometheusInterceptor object. It implements both
// the prometheus.Collector and connect.Interceptor interface.
func NewPrometheusInterceptor(counterOpts ...CounterOption) PrometheusInterceptor {
	opts := counterOptions(counterOpts)
	return &interceptor{
		serverStartedCounter: prom.NewCounterVec(
			opts.apply(prom.CounterOpts{
				Name: "grpc_server_started_total",
				Help: "Total number of RPCs started on the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledCounter: prom.NewCounterVec(
			opts.apply(prom.CounterOpts{
				Name: "grpc_server_handled_total",
				Help: "Total number of RPCs completed on the server, regardless of success or failure.",
			}), []string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"}),
		serverStreamMsgReceived: prom.NewCounterVec(
			opts.apply(prom.CounterOpts{
				Name: "grpc_server_msg_received_total",
				Help: "Total number of RPC stream messages received on the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverStreamMsgSent: prom.NewCounterVec(
			opts.apply(prom.CounterOpts{
				Name: "grpc_server_msg_sent_total",
				Help: "Total number of gRPC stream messages sent by the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledHistogramEnabled: false,
		serverHandledHistogramOpts: prom.HistogramOpts{
			Name:    "grpc_server_handling_seconds",
			Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets: prom.DefBuckets,
		},
		serverHandledHistogram: nil,
	}
}

// EnableHandlingTimeHistogram enables histograms being registered when
// registering the ServerMetrics on a Prometheus registry. Histograms can be
// expensive on Prometheus servers. It takes options to configure histogram
// options such as the defined buckets.
func (p *interceptor) EnableHandlingTimeHistogram(opts ...HistogramOption) {
	for _, o := range opts {
		o(&p.serverHandledHistogramOpts)
	}
	if !p.serverHandledHistogramEnabled {
		p.serverHandledHistogram = prom.NewHistogramVec(
			p.serverHandledHistogramOpts,
			[]string{"grpc_type", "grpc_service", "grpc_method"},
		)
	}
	p.serverHandledHistogramEnabled = true
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (p *interceptor) Describe(ch chan<- *prom.Desc) {
	p.serverStartedCounter.Describe(ch)
	p.serverHandledCounter.Describe(ch)
	p.serverStreamMsgReceived.Describe(ch)
	p.serverStreamMsgSent.Describe(ch)
	if p.serverHandledHistogramEnabled {
		p.serverHandledHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (p *interceptor) Collect(ch chan<- prom.Metric) {
	p.serverStartedCounter.Collect(ch)
	p.serverHandledCounter.Collect(ch)
	p.serverStreamMsgReceived.Collect(ch)
	p.serverStreamMsgSent.Collect(ch)
	if p.serverHandledHistogramEnabled {
		p.serverHandledHistogram.Collect(ch)
	}
}
