package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"user-api/internal/userPack"
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
	controller := userPack.NewUserController(service)

	r := gin.Default()
	r.POST("/users", controller.CreateUser)
	r.GET("/user/:id", controller.GetUser)
	r.PATCH("/user/:id", controller.UpdateUser)

	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to run server:", err)
	}
}
