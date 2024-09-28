package userPack

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id int) (*User, error)
	UpdateUser(ctx context.Context, id int, user *User) error
}

type PostgresUserRepository struct {
	db *pgx.Conn
}

func NewPostgresUserRepository(db *pgx.Conn) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	err := r.db.QueryRow(ctx, "INSERT INTO users (firstname, lastname, email, age, created) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Firstname, user.Lastname, user.Email, user.Age, user.Created).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("CreateUser: failed to insert user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, "SELECT id, firstname, lastname, email, age, created FROM users WHERE id = $1", id).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.Age, &user.Created)
	if err != nil {
		return nil, fmt.Errorf("GetUser: failed to query user with id %d: %w", id, err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, id int, user *User) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET firstname = $1, lastname = $2, email = $3, age = $4 WHERE id = $5",
		user.Firstname, user.Lastname, user.Email, user.Age, id)
	if err != nil {
		return fmt.Errorf("UpdateUser: failed to update user with id %d: %w", id, err)
	}
	return nil
}
