package actrs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"
	"github.com/hnimtadd/run/pb/v1"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
)

type (
	Server struct {
		httpServer        *http.Server
		self              *actor.PID
		runtimeManagerPID *actor.PID
		ctx               cluster.GrainContext
		responses         map[string]chan<- *pb.HTTPResponse
		store             store.Store
		cache             store.ModCacher
	}
	ServerConfig struct {
		Addr  string
		Store store.Store
	}
)

func (s *Server) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *cluster.ClusterInit:
		fmt.Println("init")
		s.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		s.self = ctx.Self()
		s.Initialize()

	case *actor.Stopped:
		s.Stop()

	case *message.RequestMessage:
		// TODO: at here there is no guarantee
		// that the runtime could be
		// initialized in time.
		// In that case, we could spaw the runtime

		runtimePID := s.RequestRuntime(msg.Request.DeploymentId, msg.Request.Runtime)
		if runtimePID == nil {
			log.Panic("failed to request a runtime PID")
			return
		}
		fmt.Println("runtime PID", runtimePID.Id)

		defer s.ctx.Request(runtimePID, msg.Request)
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

func (s *Server) RequestRuntime(deploymentID string, runtime string) *actor.PID {
	res, err := s.ctx.RequestFuture(
		s.runtimeManagerPID,
		&message.RequestRuntimeMessage{
			DeploymentID: deploymentID,
			Runtime:      runtime,
		},
		time.Second*5,
	).Result()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return res.(*actor.PID)
}

func (s *Server) Initialize() {
	s.runtimeManagerPID = s.ctx.Cluster().Get("localRuntimeManager", KindRuntimeManager)
	fmt.Println("runtime manager id", s.runtimeManagerPID.String())

	go func() {
		fmt.Printf("serving at %v...\n", s.httpServer.Addr)
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Panic("cannot close request body", err)
		}
	}()
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")
	pathParts := strings.Split(path, "/")
	fmt.Println(pathParts)
	innerURL := ""
	if len(pathParts) > 2 {
		innerURL = fmt.Sprintf("/%s", strings.Join(pathParts[2:len(pathParts)-1], "/"))
	}

	var (
		deploy   *types.Deployment
		endpoint *types.Endpoint
		err      error
		req      = utils.MakeProtoRequest(uuid.NewString())
	)
	// first param is mode
	switch pathParts[0] {
	case "preview":
		// second param is deployID
		deploy, err = s.store.GetDeploymentByID(pathParts[1])
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
			return
		}

		endpoint, err = s.store.GetEndpointByID(deploy.EndpointID.String())
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
			return
		}

	case "live":
		// second param is endpointId
		endpoint, err = s.store.GetEndpointByID(pathParts[1])
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusBadRequest, utils.MakeErrorResponse(err))
			return
		}

		if !endpoint.HasActiveDeploy() {
			_ = utils.WriteJSON(w, http.StatusBadRequest, []byte("endpoint does not have any published deploy"))
			return
		}

		deploy, err = s.store.GetDeploymentByID(endpoint.ActiveDeploymentID.String())
		if err != nil {
			_ = utils.WriteJSON(w, http.StatusNotFound, utils.MakeErrorResponse(err))
			return
		}
	default:
		_ = utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "request must be in live or preview mode"})
		return
	}

	rHeaders := make(map[string]*pb.HeaderFields)
	for k, v := range r.Header {
		field := &pb.HeaderFields{
			Fields: v,
		}
		rHeaders[k] = field
	}
	req.Env = endpoint.Environment
	req.Runtime = endpoint.Runtime
	req.Header = rHeaders
	req.EndpointId = endpoint.ID.String()
	req.DeploymentId = deploy.ID.String()
	req.Method = r.Method
	req.Url = innerURL

	bodyBuf := new(bytes.Buffer)
	_, err = io.Copy(bodyBuf, r.Body)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
		return
	}
	req.Body = bodyBuf.Bytes()

	rspCh := make(chan *pb.HTTPResponse, 1)
	reqMessage := message.NewRequestMessage(req, rspCh)
	// TODO: currently rootActor could pass the request to the server
	// consider using cluster

	if err != nil {
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.MakeErrorResponse(err))
		return
	}

	s.ctx.Send(s.self, reqMessage)
	fmt.Println("waiting for response...")
	rsp := <-rspCh

	w.WriteHeader(int(rsp.Code))
	_, _ = w.Write(rsp.Body)
}

func NewServer(cfg *ServerConfig) actor.Producer {
	return func() actor.Actor {
		s := &Server{
			responses: make(map[string]chan<- *pb.HTTPResponse),
			store:     cfg.Store,
			cache:     store.NewMemoryModCacher(),
		}
		server := &http.Server{Addr: cfg.Addr, Handler: s}
		s.httpServer = server
		return s
	}
}

// actor-related setting

var KindServer = "kind-wasm-server"

func NewServerKind(cfg *ServerConfig, opts ...actor.PropsOption) *cluster.Kind {
	return cluster.NewKind(KindServer, actor.PropsFromProducer(NewServer(cfg), opts...))
}
