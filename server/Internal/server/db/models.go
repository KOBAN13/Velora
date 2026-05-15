package db

import "context"

type User struct {
	ID           uint64
	Username     string
	PasswordHash string
}

type DbTx struct {
	Ctx            context.Context
	UserRepository *UserRepository
}
