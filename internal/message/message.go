package message

import "github.com/hnimtadd/run/pb/v1"

type Type byte
type ContextKey string

const (
	TypeRequestRuntime Type = iota
	TypeRemoveRuntime
	TypeRequest
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

type StartMessage struct{}
