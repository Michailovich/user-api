package userGrpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	userpb "user-api/gen/user"
	userPack "user-api/internal/user-pack"

	"google.golang.org/grpc"
)

type grpcServer struct {
	userpb.UnimplementedUserServiceServer
	userService *userPack.UserService
}

func NewGRPCServer(userService *userPack.UserService) *grpcServer {
	return &grpcServer{
		UnimplementedUserServiceServer: userpb.UnimplementedUserServiceServer{},
		userService:                    userService}
}

func (s *grpcServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request: %+v", req)

	if req.User == nil {
		return nil, fmt.Errorf("user data is nil")
	}

	user, err := convertProtoUserToUser(req.User)
	log.Printf("Converted user: %+v", user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proto user to user: %w", err)
	}

	if err := userPack.ValidateUser(user); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	if err := s.userService.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &userpb.CreateUserResponse{User: req.User}, nil
}

func (s *grpcServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, int(req.Id))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	protoUser := &userpb.User{
		Id:        int32(user.ID),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		Age:       uint32(user.Age),
		Created:   user.Created.Format(time.RFC3339),
	}

	return &userpb.GetUserResponse{User: protoUser}, nil
}

func (s *grpcServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	user, err := convertProtoUserToUser(req.User)

	if err != nil {
		return nil, fmt.Errorf("failed to convert proto user to user: %w", err)
	}
	if err := s.userService.UpdateUser(ctx, int(req.Id), user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &userpb.UpdateUserResponse{Message: "User updated"}, nil
}

func StartGRPCServer(userService *userPack.UserService, port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, NewGRPCServer(userService))

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	log.Printf("gRPC server is running on port %s", port)
	return nil
}

func convertProtoUserToUser(protoUser *userpb.User) (*userPack.User, error) {
	if protoUser == nil {
		return nil, fmt.Errorf("protoUser is nil")
	}

	if protoUser.Firstname == "" || protoUser.Lastname == "" || protoUser.Email == "" {
		return nil, fmt.Errorf("firstname, lastname, and email are required")
	}

	return &userPack.User{
		ID:        int(protoUser.Id),
		Firstname: protoUser.Firstname,
		Lastname:  protoUser.Lastname,
		Email:     protoUser.Email,
		Age:       uint(protoUser.Age),
		Created:   time.Now(),
	}, nil
}
