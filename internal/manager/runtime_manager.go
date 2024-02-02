package manager

import (
	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/runtime"
)

type RuntimeManager struct {
	lookup map[string]*runtime.Runtime
}

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		lookup: make(map[string]*runtime.Runtime),
	}
}

func (rm *RuntimeManager) Receive(msg *message.Message) {
	switch msg.Header {
	case message.MessageTypeRequestRuntime:
	}
}

func (rm *RuntimeManager) InitializedRuntime() {
}
