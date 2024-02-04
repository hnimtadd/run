package message

import "io"

type MessageType byte

const (
	MessageTypeRequestRuntime MessageType = iota
	MessageTypeRemoveRuntime
	MessageTypeRequest
)

type Message struct {
	Body   any
	Header MessageType
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
