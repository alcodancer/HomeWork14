package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"

	userv1 "example/api/user/v1"
	"example/internal/service"
	mongostore "example/internal/storage/mongo"
	redisstore "example/internal/storage/redis"
	grpctransport "example/internal/transport/grpc"
	httptransport "example/internal/transport/http"
)

func main() {
	ctx := context.Background()

	// --- MongoDB ---
	mongoClient, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("mongo connect:", err)
	}
	defer func() { _ = mongoClient.Disconnect(ctx) }()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("mongo ping:", err)
	}

	db := mongoClient.Database("appdb")
	userRepo := mongostore.NewUserRepo(db)
	if err := userRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal("mongo index:", err)
	}

	// --- Redis ---
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("redis ping:", err)
	}

	historyRepo := redisstore.NewHistoryRepo(redisClient, 7*24*time.Hour)

	// --- Service ---
	svc := service.NewUserService(userRepo, historyRepo)

	// --- HTTP ---
	go func() {
		h := httptransport.NewUserHandler(svc)
		mux := http.NewServeMux()
		mux.HandleFunc("/users/create", h.Create)
		mux.HandleFunc("/users/get", h.Get)
		mux.HandleFunc("/users/update", h.Update)
		mux.HandleFunc("/users/delete", h.Delete)
		mux.HandleFunc("/users/list", h.List)
		mux.HandleFunc("/users/history", h.History)

		log.Println("HTTP on :8080")
		log.Fatal(http.ListenAndServe(":8080", mux))
	}()

	// --- gRPC ---
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcSrv, grpctransport.NewUserServer(svc))

	log.Println("gRPC on :9090")
	log.Fatal(grpcSrv.Serve(lis))
}