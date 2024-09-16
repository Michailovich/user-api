package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/users", createUser)

	// Тест на отсутствие обязательных полей
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

	// Тест на неправильный формат email
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

	req, _ := http.NewRequest(http.MethodGet, "/user/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEditUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.PATCH("/user/:id", editUser)

	user := User{
		Firstname: "Jane",
		Lastname:  "Doe",
		Email:     "jane.doe@example.com",
		Age:       25,
	}

	jsonUser, _ := json.Marshal(user)
	req, _ := http.NewRequest(http.MethodPatch, "/user/1", bytes.NewBuffer(jsonUser)) // Замените на существующий ID
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func setupTestDB() {

	conn, err := pgx.Connect(context.Background(), "postgres://test_user:test_password@localhost:5433/test_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, firstname VARCHAR(100), lastname VARCHAR(100), email VARCHAR(100), age INT, created TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO users (firstname, lastname, email, age) VALUES ('Test', 'User', 'test.user@example.com', 30)")
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	setupTestDB()
	code := m.Run()
	os.Exit(code)
}
