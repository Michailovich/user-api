package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type User struct {
	ID        int       `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	Age       uint      `json:"age"`
	Created   time.Time `json:"created"`
}

var db *pgx.Conn

func initDB() {
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

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func validateUser(user User) error {
	if user.Firstname == "" || user.Lastname == "" || user.Email == "" {
		return fmt.Errorf("firstname, lastname and email are required")
	}
	if !isValidEmail(user.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.Created = time.Now()

	if err := validateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.QueryRow(context.Background(), "INSERT INTO users (firstname, lastname, email, age, created) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Firstname, user.Lastname, user.Email, user.Age, user.Created).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func getUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user User
	err = db.QueryRow(context.Background(), "SELECT id, firstname, lastname, email, age, created FROM users WHERE id = $1", id).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.Age, &user.Created)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func editUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec(context.Background(), "UPDATE users SET firstname = $1, lastname = $2, email = $3, age = $4 WHERE id = $5",
		user.Firstname, user.Lastname, user.Email, user.Age, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func main() {
	initDB()
	defer db.Close(context.Background())

	r := gin.Default()
	r.POST("/users", createUser)
	r.GET("/user/:id", getUser)
	r.PATCH("/user/:id", editUser)

	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to run server:", err)
	}
}
