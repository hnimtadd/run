package manager

import (
	"log"
	"net/http"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/pb/v1"
)

type Server struct {
	server   *http.Server
	rm       *RuntimeManager
	response map[string]chan<- *pb.HTTPResponse
}

func NewServer(server *http.Server) *Server {
	s := &Server{rm: NewRuntimeManager(), response: make(map[string]chan<- *pb.HTTPResponse), server: server}
	s.initialize()
	return s
}

func (s *Server) ReceiveMessage(m any) {
	switch msg := m.(type) {
	case *message.StartMessage:
		s.initialize()
	// Receive request message with response function
	case *message.RequestMessage:
		go s.rm.Receive(&message.Message{Header: message.TypeRequest, Body: msg})

	case *pb.HTTPResponse:
		msgID := msg.RequestId
		rspCh, ok := s.response[msgID]
		if !ok {
			return
		}
		rspCh <- msg
	}
}

func (s *Server) initialize() {
	go func() {
		log.Fatal(s.server.ListenAndServe())
	}()
}
