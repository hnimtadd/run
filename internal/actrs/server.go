package actrs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/asynkron/protoactor-go/actor"
)

type Server struct {
	httpServer        *http.Server
	self              *actor.PID
	runtimeManagerPID *actor.PID
	responses         map[string]chan<- *pb.HTTPResponse
	store             store.Store
	cache             store.ModCacher
}

func (s *Server) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		s.self = ctx.Self()
		s.Initialize()
		runtimeManagerPID := ctx.Spawn(actor.PropsFromProducer(NewRuntimeManager(s.store, s.cache)))
		s.runtimeManagerPID = runtimeManagerPID

	case *actor.Stopped:
		s.Stop()

	case *message.RequestMessage:

		// TODO: at here there is no guarantee
		// that the runtime could be
		// initialized in time.
		// In that case, we could spaw the runtime

		runtimePID := s.RequestRuntime(ctx, msg.Request.DeploymentId, msg.Request.Runtime)
		if runtimePID == nil {
			log.Panic("failed to request a runtime PID")
			return
		}
		defer ctx.RequestWithCustomSender(runtimePID, msg, s.self)
		if msg.ResponseCh == nil {
			return
		}

		s.responses[msg.Request.Id] = msg.ResponseCh

	case *pb.HTTPResponse:
		if responseCh, ok := s.responses[msg.RequestId]; ok {
			responseCh <- msg
			delete(s.responses, msg.RequestId)
		}
	}
}

func (s *Server) RequestRuntime(c actor.Context, deploymentID string, runtime string) *actor.PID {
	res, err := c.RequestFuture(
		s.runtimeManagerPID,
		message.RequestRuntimeMessage{
			DeploymentID: deploymentID,
			Runtime:      runtime,
		},
		time.Second*2,
	).Result()
	if err != nil {
		return nil
	}
	return res.(*actor.PID)
}

func (s *Server) Initialize() {
	fmt.Println("serving...")
	go func() {
		log.Panic(s.httpServer.ListenAndServe())
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("cannot shutdown server, %v", err)
	}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("hello world"))
	if err != nil {
		panic(err)
	}
}

func NewServer(addr string) actor.Producer {
	return func() actor.Actor {
		s := &Server{
			responses: make(map[string]chan<- *pb.HTTPResponse),
			store:     store.NewMemoryStore(),
			cache:     store.NewMemoryModCacher(),
		}
		server := &http.Server{Addr: addr, Handler: s}
		s.httpServer = server
		return s
	}
}
