package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userPack "user-api/internal/user-pack"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *userPack.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUser(ctx context.Context, id int) (*userPack.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*userPack.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, id int, user *userPack.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepository)
	userService := userPack.NewUserService(mockRepo)
	handler := userPack.NewUserHandler(userService)

	router.POST("/users", handler.CreateUser)

	// Test case: Missing Firstname
	user := userPack.User{
		Lastname: "Doe",
		Email:    "john.doe@example.com",
		Age:      30,
	}

	jsonUser, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonUser))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case: Invalid Email
	user = userPack.User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "invalid-email",
		Age:       30,
	}

	jsonUser, err = json.Marshal(user)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonUser))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case: Successful User Creation
	user = userPack.User{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *userPack.User) bool {
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
	userService := userPack.NewUserService(mockRepo)
	handler := userPack.NewUserHandler(userService)

	router.GET("/user/:id", handler.GetUser)

	testUser := &userPack.User{
		ID:        1,
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
		Age:       30,
	}

	// Test case: User found
	mockRepo.On("GetUser", mock.Anything, 1).Return(testUser, nil)

	req, err := http.NewRequest(http.MethodGet, "/user/1", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	// Test case: User not found
	mockRepo.On("GetUser", mock.Anything, 2).Return((*userPack.User)(nil), errors.New("user not found"))

	req, err = http.NewRequest(http.MethodGet, "/user/2", nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestEditUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepository)
	userService := userPack.NewUserService(mockRepo)
	handler := userPack.NewUserHandler(userService)

	router.PATCH("/user/:id", handler.UpdateUser)

	// Test case: Successful User Update
	updatedUser := userPack.User{
		Firstname: "Jane",
		Lastname:  "Doe",
		Email:     "jane.doe@example.com",
		Age:       25,
	}

	mockRepo.On("UpdateUser", mock.Anything, 1, &updatedUser).Return(nil)

	jsonUser, err := json.Marshal(updatedUser)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, "/user/1", bytes.NewBuffer(jsonUser))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	// Test case: User not found
	mockRepo.On("UpdateUser", mock.Anything, 2, &updatedUser).Return(errors.New("user not found"))

	req, err = http.NewRequest(http.MethodPatch, "/user/2", bytes.NewBuffer(jsonUser))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}
