package manager

import (
	"context"
	"log"

	"github.com/hnimtadd/run/internal/message"
)

// wasm server->runtime manager->runtime ??
// actor:separate instance without call directly with code

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
	case message.TypeRequestRuntime:
		requestRuntimeMessage, ok := msg.Body.(*message.RequestRuntimeMessage)
		if !ok {
			log.Panic("unexpected error")
		}
		runtime := NewRuntime()
		if err := runtime.Initialize(requestRuntimeMessage); err != nil {
			log.Panic(err)
		}
		rm.lookup[requestRuntimeMessage.DeploymentID] = runtime

	case message.TypeRemoveRuntime:
		removeRuntimeMessage, ok := msg.Body.(*message.RemoveRuntimeMessage)
		if !ok {
			log.Panic("unexpected error")
			return
		}
		_, ok = rm.lookup[removeRuntimeMessage.DeploymentID]
		if ok {
			delete(rm.lookup, removeRuntimeMessage.DeploymentID)
		}

	case message.TypeRequest:
		msg, ok := msg.Body.(*message.RequestMessage)
		if !ok {
			log.Panic("unexpected error")
		}
		runtime, ok := rm.lookup[msg.Request.DeploymentId]
		if !ok {
			return
		}
		ctx := context.WithValue(context.Background(), message.ContextKey(msg.Request.Id), msg.ResponseCh)
		runtime.Handle(ctx, msg.Request)
	}
}
