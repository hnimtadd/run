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
		if err := runtime.Initialize(requestRuntimeMessage); err != nil {
			log.Panic(err)
		}
		rm.lookup[requestRuntimeMessage.DeploymentID] = runtime
	case message.MessageTypeRemoveRuntime:
		removeRuntimeMessage, ok := msg.Body.(*message.RemoveRuntimeMessage)
		if !ok {
			log.Panic("unexpected error")
			return
		}
		_, ok = rm.lookup[removeRuntimeMessage.DeploymentID]
		if ok {
			delete(rm.lookup, removeRuntimeMessage.DeploymentID)
		}
	case message.MessageTypeRequest:
		requestMessage, ok := msg.Body.(*message.RequestMessage)
		if !ok {
			log.Panic("unexpected error")
		}
		runtime, ok := rm.lookup[requestMessage.DeploymentID]
		if !ok {
			return
		}
		if err := runtime.Handle(requestMessage); err != nil {
			log.Panic(err)
		}
	}
}
