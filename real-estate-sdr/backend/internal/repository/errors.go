package repository

import "errors"

var (
	ErrLeadNotFound      = errors.New("lead not found")
	ErrLeadAlreadyExists = errors.New("lead already exists")
)
