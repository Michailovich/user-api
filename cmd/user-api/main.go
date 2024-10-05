package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	server "user-api/internal/grpc/user"
	userPack "user-api/internal/user-pack"
	"user-api/pkg/db"
)

func main() {
	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer db.Close(context.Background())

	repo := userPack.NewPostgresUserRepository(db)
	service := userPack.NewUserService(repo)
	handler := userPack.NewUserHandler(service)

	r := gin.Default()
	r.POST("/users", handler.CreateUser)
	r.GET("/user/:id", handler.GetUser)
	r.PATCH("/user/:id", handler.UpdateUser)

	go func() {
		if err := r.Run(":8080"); err != nil {
			fmt.Println("Failed to run server:", err)
		}
	}()

	go func() {
		if err := server.StartGRPCServer(service, ":50051"); err != nil {
			log.Fatalf("failed to start gRPC server: %v", err)
		}
		log.Printf("gRPC server is running on port %s", ":50051")
	}()

	select {}
}
