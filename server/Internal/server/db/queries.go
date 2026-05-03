package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbPool *pgxpool.Pool
}

type CreateUserParams struct {
	Username     string
	PasswordHash string
}

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrUsersTableNotProvided = errors.New("users table name is not provided")
)

func usersTableName() (string, error) {
	var usersTable = os.Getenv("USERS_TABLE")

	if usersTable == "" {
		return "", ErrUsersTableNotProvided
	}

	return pgx.Identifier{usersTable}.Sanitize(), nil
}

func TestPostgresConnection() {
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	var pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Error create pool: %s", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Error db ping: %s", err)
	}

	var result int

	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		log.Fatalf("Error querying row: %s", err)
	}

	if result != 1 {
		log.Fatalf("Expected 1 row, got %d", result)
	}

	log.Println("Successfully connected to database")
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{dbPool: dbPool}
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var tableName, tableNameErr = usersTableName()

	if tableNameErr != nil {
		return nil, tableNameErr
	}

	var sqlRequest = fmt.Sprintf(`SELECT id, username, password_hash FROM %s WHERE username = $1`, tableName)

	var user User

	var err = r.dbPool.QueryRow(ctx, sqlRequest, username).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	var tableName, tableNameErr = usersTableName()
	if tableNameErr != nil {
		return nil, tableNameErr
	}

	var query = fmt.Sprintf(`INSERT INTO %s (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash`, tableName)

	var user User

	var err = r.dbPool.QueryRow(ctx, query, params.Username, params.PasswordHash).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}

		return nil, err
	}

	return &user, nil
}
