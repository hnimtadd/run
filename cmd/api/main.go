/*
api cmd boot the api server which serve rest apis.
*/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hnimtadd/run/internal/api"
	"github.com/hnimtadd/run/internal/store"
	"github.com/hnimtadd/run/internal/version"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	// TODO: implement logging store, currently use in-memory store, use in-memory store make no sense here.
	inMemoryStore := store.NewMemoryStore()

	// TODO: integrate blob store
	creds := credentials.NewStaticV4(os.Getenv("MINIO_USERNAME"), os.Getenv("MINIO_PASSWORD"), "")
	minioClient, err := minio.New(os.Getenv("MINIO_URL"), &minio.Options{
		Creds:  creds,
		Secure: false,
	})
	if err != nil {
		slog.Error("cannot init minio client", "msg", err.Error())
	}

	blobStore, err := store.NewMinioBlobStore(minioClient, os.Getenv("MINIO_BUCKET"))
	if err != nil {
		slog.Error("cannot init blob store", "msg", err.Error())
	}

	serverConfig := api.ServerConfig{
		Addr:    fmt.Sprintf(":%v", os.Getenv("API_ADDR")),
		Version: version.Version,
	}
	apiServer := api.NewServer(st, inMemoryStore, blobStore, serverConfig)

	go func() {
		panic(apiServer.ListenAndServe())
	}()

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	signal.Notify(exitCh, syscall.SIGTERM)
	<-exitCh
}
