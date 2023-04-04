package gorequests

import (
	"context"
	"fmt"
	"net/http"
)

type LogProducer interface {
	SendLogMessage(ctx context.Context, data []byte) (string, error)
}

// ==================

func NewDiscardLogProducer() LogProducer {
	return newDiscardLogProducer()
}

func NewPrinterLogProducer() LogProducer {
	return newPrinterLogProducer()
}

func NewRmqLogProducer(producer Producer, topic string) LogProducer {
	return newRmqLogProducer(producer, topic)
}

// discard log producer
type discardLogProducer struct {
}

func newDiscardLogProducer() LogProducer {
	return &discardLogProducer{}
}

func (p *discardLogProducer) SendLogMessage(ctx context.Context, data []byte) (string, error) {
	return "", nil
}

// printer log producer
type printerLogProducer struct {
}

func newPrinterLogProducer() LogProducer {
	return &printerLogProducer{}
}

func (p *printerLogProducer) SendLogMessage(ctx context.Context, data []byte) (string, error) {
	fmt.Printf("[PrinterLogProducer] send log message: %s\n", string(data))
	return "", nil
}

type Producer interface {
	Send(ctx context.Context, data []byte) (string, error)
}

// rocket mq log producer
type rmqLogProducer struct {
	producer Producer
	topic    string
}

func newRmqLogProducer(producer Producer, topic string) LogProducer {
	return &rmqLogProducer{
		producer: producer,
		topic:    topic,
	}
}

func (p *rmqLogProducer) SendLogMessage(ctx context.Context, data []byte) (string, error) {
	return p.producer.Send(ctx, data)
}

type RequestMessageType int

const (
	RequestMessageTypeUnknown RequestMessageType = 0
	RequestMessageTypeIn      RequestMessageType = 1
	RequestMessageTypeOut     RequestMessageType = 2
)

type LogMessage struct {
	Method string `json:"method"`
	Url    string `json:"url"`

	RequestBody   string      `json:"request_body"`
	RequestHeader http.Header `json:"request_header"`
	RequestTime   string      `json:"request_time"`

	ResponseBody      string      `json:"response_body"`
	ResponseHeader    http.Header `json:"response_header"`
	ResponseStateCode int         `json:"response_state_code"`
	ResponseTime      string      `json:"response_time"`

	TimeConsuming int64  `json:"time_consuming"` // milliseconds
	ErrorMessage  string `json:"error_message"`

	LogId       string             `json:"log_id"`
	RequestType RequestMessageType `json:"request_type"`
}
