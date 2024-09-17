package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	os.Setenv("DATABASE_URL", "postgres://user:password@localhost:5433/testdb?sslmode=disable")

	initDBTest()
	defer db.Close(context.Background())

	exitVal := m.Run()

	os.Exit(exitVal)
}

func initDBTest() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	db, err = pgx.Connect(context.Background(), dbURL)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            firstname VARCHAR(100) NOT NULL,
            lastname VARCHAR(100) NOT NULL,
            email VARCHAR(100) NOT NULL UNIQUE,
            age INT,
            created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}
}
func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/users", createUser)

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	createUserWithTx := func(c *gin.Context) {
		c.Set("db", tx)
		createUser(c)
	}

	router.POST("/users", createUserWithTx)

	user := User{
		Lastname: "Doe",
		Email:    "john.doe@example.com",
		Age:      30,
	}

	jsonUser, _ := json.Marshal(user)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	user = User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "invalid-email",
		Age:       30,
	}

	jsonUser, _ = json.Marshal(user)
	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/user/:id", getUser)

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	getUserWithTx := func(c *gin.Context) {
		c.Set("db", tx)
		getUser(c)
	}

	router.GET("/user/:id", getUserWithTx)

	testUser := User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO users (firstname, lastname, email, age) VALUES ($1, $2, $3, $4)",
		testUser.Firstname, testUser.Lastname, testUser.Email, testUser.Age)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/user/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEditUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.PATCH("/user/:id", editUser)

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	editUserWithTx := func(c *gin.Context) {
		c.Set("db", tx)
		editUser(c)
	}

	router.PATCH("/user/:id", editUserWithTx)

	testUser := User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO users (firstname, lastname, email, age) VALUES ($1, $2, $3, $4)",
		testUser.Firstname, testUser.Lastname, testUser.Email, testUser.Age)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	updatedUser := User{
		Firstname: "Jane",
		Lastname:  "Doe",
		Email:     "jane.doe@example.com",
		Age:       25,
	}

	jsonUser, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest(http.MethodPatch, "/user/1", bytes.NewBuffer(jsonUser)) // Замените на существующий ID
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
