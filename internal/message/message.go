package message

type MessageType byte

const (
	MessageTypeRequestRuntime MessageType = iota
	MessageTypeRemoveRuntime
)

type Message struct {
	Body   any
	Header MessageType
}

type RequestRuntimeMessage struct {
	DeploymentID string
	Runtime      string
}
