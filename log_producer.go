package gorequests

import (
	"code.byted.org/rocketmq/rocketmq-go-proxy/pkg/producer"
	"code.byted.org/rocketmq/rocketmq-go-proxy/pkg/types"
	"context"
	"fmt"
	"net/http"
	"time"
)

var (
	logProducer LogProducer
)

type LogProducer interface {
	SendLogMessage(ctx context.Context, data []byte) error
}

// ==================

func NewDiscardLogProducer() LogProducer {
	return newDiscardLogProducer()
}

func NewPrinterLogProducer() LogProducer {
	return newPrinterLogProducer()
}

func NewRmqLogProducer(producer producer.Producer, topic string) LogProducer {
	return newRmqLogProducer(producer, topic)
}

// discard log producer
type discardLogProducer struct {
}

func newDiscardLogProducer() LogProducer {
	return &discardLogProducer{}
}

func (p *discardLogProducer) SendLogMessage(ctx context.Context, data []byte) error {
	return nil
}

// printer log producer
type printerLogProducer struct {
}

func newPrinterLogProducer() LogProducer {
	return &printerLogProducer{}
}

func (p *printerLogProducer) SendLogMessage(ctx context.Context, data []byte) error {
	fmt.Printf("[PrinterLogProducer] send log message: %s\n", string(data))
	return nil
}

// rocket mq log producer
type rmqLogProducer struct {
	producer producer.Producer
	topic    string
}

func newRmqLogProducer(producer producer.Producer, topic string) LogProducer {
	return &rmqLogProducer{
		producer: producer,
		topic:    topic,
	}
}

func (p *rmqLogProducer) SendLogMessage(ctx context.Context, data []byte) error {
	msg := types.NewMessage(p.topic, data)
	_, err := p.producer.Send(ctx, msg)
	return err
}

type RequestMessageType int

const (
	RequestMessageType_Unknown RequestMessageType = 0
	RequestMessageType_In      RequestMessageType = 1
	RequestMessageType_Out     RequestMessageType = 2
)

type LogMessage struct {
	Method string `json:"method"`
	Url    string `json:"url"`

	RequestBody   string      `json:"request_body"`
	RequestHeader http.Header `json:"request_header"`
	RequestTime   time.Time   `json:"request_time"`

	ResponseBody      string      `json:"response_body"`
	ResponseHeader    http.Header `json:"response_header"`
	ResponseStateCode int         `json:"response_state_code"`
	ResponseTime      time.Time   `json:"response_time"`

	TimeConsuming int64  `json:"time_consuming"` // milliseconds
	ErrorMessage  string `json:"error_message"`

	LogId       string             `json:"log_id"`
	RequestType RequestMessageType `json:"request_type"`
}
