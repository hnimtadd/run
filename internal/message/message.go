package message

import (
	"github.com/hnimtadd/run/internal/types"

	"github.com/hnimtadd/run/pb/v1"
)

type (
	Type       byte
	ContextKey string
)

const (
	TypeRequestRuntime Type = iota
	TypeRemoveRuntime
	TypeRequest
	TypeMetric
	TypeRuntimeResponse
)

type Message struct {
	Body   any
	Header Type
}

type RequestRuntimeMessage struct {
	DeploymentID string
	Runtime      string
}

type RemoveRuntimeMessage struct {
	DeploymentID string
}

type RequestMessage struct {
	Request    *pb.HTTPRequest
	ResponseCh chan<- *pb.HTTPResponse
}

func NewRequestMessage(req *pb.HTTPRequest, rspCh chan<- *pb.HTTPResponse) *RequestMessage {
	return &RequestMessage{
		Request:    req,
		ResponseCh: rspCh,
	}
}

type StartMessage struct{}

type MetricMessage struct {
	DeploymentID string // uuid.UUID
	RequestID    string // uuid.UUID
	Metric       types.RequestMetric
}

type ResponseWithMetric struct {
	Response      *pb.HTTPResponse
	MetricMessage *MetricMessage
}
