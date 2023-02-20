package prometheus

import (
	"github.com/bufbuild/connect-go"
)

type monitoringHandler struct {
	connect.StreamingHandlerConn
	reporter
}

func (m monitoringHandler) Receive(msg any) error {
	err := m.StreamingHandlerConn.Receive(msg)
	if err == nil {
		m.reporter.monitorReceive()
	}
	return err
}

func (m monitoringHandler) Send(msg any) error {
	err := m.StreamingHandlerConn.Send(msg)
	if err == nil {
		m.reporter.monitorSend()
	}
	return err
}
