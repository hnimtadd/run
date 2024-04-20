/*
ingress cmd boot the ingress server which redirect the request to the relevant runtime.
*/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hnimtadd/run/internal/actrs"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/version"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/automanaged"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	envFile := ".env"
	if env != "" {
		envFile = fmt.Sprintf("%s.%s", envFile, env)
	}
	if err := godotenv.Load(envFile); err != nil {
		slog.Error("could not load envFile", "at", envFile, "msg", err.Error())
		return
	}

	// init mongo store
	url := os.Getenv("MONGO_URL")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	opt := options.Client().ApplyURI(url)
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		slog.Error("could not connect to mongo db", "msg", err.Error())
		return
	}
	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("could not ping to mongo server", "msg", err.Error())
		return
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			slog.Error("failed to disconnect to mongo, force disconnect")
		}
	}()

	db := client.Database(os.Getenv("MONGO_DATABASE"))
	st, err := store.NewMongoStore(db)
	if err != nil {
		slog.Error("cannot init store with given mongo client", "msg", err.Error())
	}

	// TODO: implement logging store and cache store, currently use in-memory store.
	inMemoryStore := store.NewMemoryStore()
	inMemoryCache := store.NewMemoryModCacher()
	// TODO: integrate blob store
	creds := credentials.NewStaticV4(
		os.Getenv("MINIO_USERNAME"),
		os.Getenv("MINIO_PASSWORD"),
		"",
	)
	minioClient, err := minio.New(
		os.Getenv("MINIO_URL"),
		&minio.Options{
			Creds:  creds,
			Secure: false,
		},
	)
	if err != nil {
		slog.Error("cannot init minio client", "msg", err.Error())
	}

	blobStore, err := store.NewMinioBlobStore(
		minioClient,
		os.Getenv("MINIO_BUCKET"),
	)
	if err != nil {
		slog.Error("cannot init blob store", "msg", err.Error())
	}

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
					Addr:    fmt.Sprintf(":%v", os.Getenv("WASM_ADDR")),
					Store:   st,
					Version: version.Version,
				}),
			actrs.NewRuntimeManagerKind(),
			actrs.NewRuntimeKind(
				&actrs.RuntimeConfig{
					Store:     st,
					Cache:     inMemoryCache,
					LogStore:  inMemoryStore,
					BlobStore: blobStore,
				}),
			actrs.NewMetricAggregatorKind(
				&actrs.MetricAggregatorConfig{
					Version:       version.Version,
					MetricStore:   nil, // TODO: implement this
					MetadataStore: st,
				},
			),
		),
	)

	c := cluster.New(system, clusterConfig)
	c.StartMember()

	pid := c.Get(os.Getenv("ACTOR_WASM_SERVER_ID"), actrs.KindServer)
	fmt.Println(pid)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	signal.Notify(exitCh, syscall.SIGTERM)
	<-exitCh
}
