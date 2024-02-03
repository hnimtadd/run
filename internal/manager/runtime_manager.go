package manager

import (
	"log"

	"github.com/hnimtadd/run/internal/message"
)

type RuntimeManager struct {
	lookup map[string]*Runtime
}

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		lookup: make(map[string]*Runtime),
	}
}

func (rm *RuntimeManager) Receive(msg *message.Message) {
	switch msg.Header {
	case message.MessageTypeRequestRuntime:
		requestRuntimeMessage, ok := msg.Body.(*message.RequestRuntimeMessage)
		if !ok {
			log.Panic("unexpected error")
		}
		runtime := NewRuntime()
		runtime.Initialize(requestRuntimeMessage)
		rm.lookup[requestRuntimeMessage.DeploymentID] = runtime
	}
}
