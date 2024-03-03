package message

import "io"

type Type byte

const (
	MessageTypeRequestRuntime Type = iota
	MessageTypeRemoveRuntime
	MessageTypeRequest
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
	DeploymentID string
	Body         io.Reader
	Env          map[string]string
	Args         []string
}
