package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hnimtadd/run/internal/actrs"
	"github.com/hnimtadd/run/internal/api"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/types"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var c *cluster.Cluster

func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
}

func main() {
	st := store.NewMemoryStore()
	MockStore(st)
	mc := store.NewMemoryModCacher()

	system := actor.NewActorSystem()
	defer system.Shutdown()
	provider := automanaged.New()
	lookup := disthash.New()
	clusterPort, err := strconv.Atoi(os.Getenv("ACTOR_CLUSTER_PORT"))
	if err != nil {
		panic(err)
	}
	config := remote.Configure(os.Getenv("ACTOR_CLUSTER_HOST"), clusterPort)
	clusterConfig := cluster.Configure(
		os.Getenv("ACTOR_CLUSTER_NAME"),
		provider,
		lookup,
		config,
		cluster.WithKinds(
			actrs.NewServerKind(
				&actrs.ServerConfig{
					Addr:  fmt.Sprintf(":%v", os.Getenv("WASM_ADDR")),
					Store: st,
				}),
			actrs.NewRuntimeManagerKind(),
			actrs.NewRuntimeKind(
				&actrs.RuntimeConfig{
					Store:    st,
					Cache:    mc,
					LogStore: st,
				}),
		),
	)

	c = cluster.New(system, clusterConfig)
	c.StartMember()

	pid := c.Get(os.Getenv("ACTOR_WASM_SERVER_ID"), actrs.KindServer)
	fmt.Println(pid)

	apiServer := api.NewServer(st, st)
	go func() {
		panic(apiServer.ListenAndServe(fmt.Sprintf(":%v", os.Getenv("API_ADDR"))))
	}()
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	signal.Notify(exitCh, syscall.SIGTERM)
	<-exitCh
}

func MockStore(s *store.MemoryStore) {
	endpoint := &types.Endpoint{
		ID:      uuid.MustParse("c495fdf0-f8c9-41ad-befd-ff9af5d0d33d"),
		Name:    "mockEndpoint",
		Runtime: "go",
		Environment: map[string]string{
			"my_name": "my_name",
		},
	}
	err := s.CreateEndpoint(endpoint)
	if err != nil {
		panic(err)
	}

	var blob []byte
	blob, err = os.ReadFile("./examples/go/example.wasm")
	if err != nil {
		panic(err)
	}
	var deployment *types.Deployment
	deployment, err = types.NewDeployment(endpoint, blob)
	if err != nil {
		panic(err)
	}
	err = s.CreateDeployment(deployment)
	if err != nil {
		panic(err)
	}
}
