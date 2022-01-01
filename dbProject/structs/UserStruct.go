package structs

import "github.com/jackc/pgtype"

type User struct {
	ID        string             `db:"id" json:"id"`
	Login     string             `db:"login" json:"login"`
	Password  string             `db:"password" json:"password"`
	CreatedAt pgtype.Timestamptz `db:"created_at" json:"created_at"`
}

type UserFromBody struct {
	Login    string `db:"login" json:"login"`
	Password string `db:"password" json:"password"`
}

type TokenResp struct {
	Token string `json:"token"`
}

type CustomError struct {
	Message string `json:"message"`
}