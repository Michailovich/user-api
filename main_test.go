package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1) // Return nil if the user is not found
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, id int, user *User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	controller := NewUserController(userService)

	router.POST("/users", controller.CreateUser)

	// Test case: Missing Firstname
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

	// Test case: Invalid Email
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

	// Test case: Successful User Creation
	user = User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
		return u.Firstname == "John" && u.Lastname == "Doe" && u.Email == "john.doe@example.com" && u.Age == 30
	})).Return(nil)

	jsonUser, _ = json.Marshal(user)
	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	controller := NewUserController(userService)

	router.GET("/user/:id", controller.GetUser)

	testUser := &User{
		ID:        1,
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	// Test case: User found
	mockRepo.On("GetUserByID", mock.Anything, 1).Return(testUser, nil)

	req, _ := http.NewRequest(http.MethodGet, "/user/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	// Test case: User not found
	mockRepo.On("GetUserByID", mock.Anything, 2).Return((*User)(nil), errors.New("user not found"))

	req, _ = http.NewRequest(http.MethodGet, "/user/2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestEditUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)
	controller := NewUserController(userService)

	router.PATCH("/user/:id", controller.UpdateUser)

	// Test case: Successful User Update
	updatedUser := User{
		Firstname: "Jane",
		Lastname:  "Doe",
		Email:     "jane.doe@example.com",
		Age:       25,
	}

	mockRepo.On("UpdateUser", mock.Anything, 1, &updatedUser).Return(nil)

	jsonUser, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest(http.MethodPatch, "/user/1", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	// Test case: User not found
	mockRepo.On("UpdateUser", mock.Anything, 2, &updatedUser).Return(errors.New("user not found"))

	req, _ = http.NewRequest(http.MethodPatch, "/user/2", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}
