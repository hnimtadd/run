package actrs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"
	"github.com/hnimtadd/run/internal/utils"
	pb "github.com/hnimtadd/run/pbs/gopb/v1"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
)

type (
	Server struct {
		httpServer          *http.Server
		self                *actor.PID
		runtimeManagerPID   *actor.PID
		metricAggregatorPID *actor.PID
		ctx                 cluster.GrainContext
		responses           map[string]chan<- *pb.HTTPResponse
		store               store.Store
		cache               store.ModCacher
		version             string
	}
	ServerConfig struct {
		Addr    string
		Store   store.Store
		Version string
	}
)

func (s *Server) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *cluster.ClusterInit:
		slog.Info("received cluster init message", "node", "Server")
		s.ctx = cluster.NewGrainContext(ctx, msg.Identity, msg.Cluster)
		s.self = ctx.Self()
		s.Initialize()

	case *actor.Stopped:
		s.Stop()

	case *message.RequestMessage:
		// at here there is no guarantee
		// that the runtime could be
		// initialized in time.
		// In that case, we could spaw the runtime

		runtimePID, err := s.RequestRuntime(msg.Request.DeploymentId, msg.Request.Runtime)
		if err != nil {
			slog.Info("cannot request runtime", "msg", err, "node", "server")
			return
		}
		slog.Info("request runtime success, redirecting user request", "pid", runtimePID.Id)

		defer s.ctx.Request(runtimePID, msg.Request)
		if msg.ResponseCh == nil {
			return
		}
		s.responses[msg.Request.Id] = msg.ResponseCh

	case *message.ResponseWithMetric:
		rsp := msg.Response
		if responseCh, ok := s.responses[rsp.RequestId]; ok {
			responseCh <- rsp
			delete(s.responses, rsp.RequestId)
		}
		if msg.MetricMessage != nil {
			ctx.Send(s.metricAggregatorPID, msg.MetricMessage)
		}
	}
}

func (s *Server) RequestRuntime(deploymentID string, runtime string) (*actor.PID, error) {
	res, err := s.ctx.RequestFuture(
		s.runtimeManagerPID,
		&message.RequestRuntimeMessage{
			DeploymentID: deploymentID,
			Runtime:      runtime,
		},
		time.Second*5,
	).Result()
	if err != nil {
		return nil, err
	}
	return res.(*actor.PID), nil
}

func (s *Server) Initialize() {
	s.runtimeManagerPID = s.ctx.Cluster().Get("localRuntimeManager", KindRuntimeManager)
	slog.Info("initialized runtime manager", "pid", s.runtimeManagerPID.Id)

	s.metricAggregatorPID = s.ctx.Cluster().Get("localMetricAggragator", KindMetricAggregator)
	slog.Info("initialized metric aggregator", "pid", s.metricAggregatorPID.Id)

	go func() {
		slog.Info("serving ingress...", "at", s.httpServer.Addr, "node", "Server", "version", s.version)
		log.Panic(s.httpServer.ListenAndServe())
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("cannot shutdown server", "msg", err.Error())
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("failed to close this request's body", "msg", err.Error())
		}
	}()
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")
	pathParts := strings.Split(path, "/")
	innerURL := "/"
	if len(pathParts) > 2 {
		innerURL = fmt.Sprintf("/%s", strings.Join(pathParts[2:], "/"))
	}

	var (
		deploy   *types.Deployment
		endpoint *types.Endpoint
		err      error
		req      = utils.MakeProtoRequest(uuid.NewString())
		mode     = pathParts[0]
	)
	slog.Info("new request", "node", "server", "environment", pathParts[0], "url", innerURL)

	// first param is mode
	switch mode {
	case "preview":
		deploymentID := pathParts[1]
		// second param is deployID
		deploy, err = s.store.GetDeploymentByID(deploymentID)
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
		endpointID := pathParts[1]
		endpoint, err = s.store.GetEndpointByID(endpointID)
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

	protoHeader := make(map[string]*pb.HeaderFields)
	for k, v := range r.Header {
		field := &pb.HeaderFields{
			Fields: v,
		}
		protoHeader[k] = field
	}

	req.Env = endpoint.Environment
	req.Runtime = endpoint.Runtime
	req.Header = protoHeader
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

	s.ctx.Send(s.self, reqMessage)
	slog.Info("waiting for response from sandbox...")
	rsp := <-rspCh

	w.WriteHeader(int(rsp.Code))
	for key, val := range rsp.Header {
		for _, field := range val.Fields {
			fmt.Println(key, field)
			w.Header().Add(key, field)
		}
	}

	slog.Info("got response from sandbox, returning to user")
	_, _ = w.Write(rsp.Body)
}

func NewServer(cfg *ServerConfig) actor.Producer {
	return func() actor.Actor {
		s := &Server{
			responses: make(map[string]chan<- *pb.HTTPResponse),
			store:     cfg.Store,
			cache:     store.NewMemoryModCacher(),
			version:   cfg.Version,
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
