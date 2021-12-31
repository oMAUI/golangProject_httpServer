package SqlCommand

import (

	"dbProject/structs"
	"errors"

	"github.com/jackc/pgtype"
	"golang.org/x/crypto/bcrypt"


	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"os"
	"time"
)

var(
	defUri = "postgresql://habitov:habitov@95.217.232.188:7777/habitov"
	localDb = "postgresql://selectel:selectel@192.168.3.30:5432/"

	ErrCompareHashAndPassword = fmt.Errorf("WrongPassword")
)

type DB struct{
	conn *sqlx.DB
}

func Connection(uri string) (*DB, error){
	connUri := uri //"postgresql://selectel:selectel@192.168.3.30:5432/"

	dataBase, err := sqlx.Connect("pgx", connUri)
	if err != nil {
		return nil, err
	}

	dataBase.SetMaxOpenConns(1)
	dataBase.SetMaxIdleConns(3)

	return &DB{
		conn: dataBase,
	}, nil
}

func (dataBase *DB) CreateUser(user structs.UserFromBody) (structs.User, error){
	var id string

	hash, errHash := HashPassword(user.Password)
	if errHash != nil {
		fmt.Println("failed to generate hash %w", errHash.Error())
		return structs.User{}, errHash
	}
	
	if errCreate := dataBase.conn.Get(&id,
		`INSERT INTO users(login, password)
				VALUES ($1, $2)
				RETURNING id`, user.Login, hash);
	errCreate != nil{
		fmt.Println("failed to create user %w", errCreate.Error())
		return structs.User{}, errCreate
	}

	User := structs.User{
		ID: id,
		Login: user.Login,
		Password: hash,
		CreatedAt: pgtype.Timestamptz{
			Time: time.Now(),
		},
	}

	return User, nil
}

func (dataBase *DB) AuthorizationUser(userData structs.UserFromBody) (structs.User, error){
	var hash string

	if err := dataBase.conn.Get(&hash, "SELECT password FROM users WHERE login = $1",
		userData.Login); err != nil {
		//fmt.Println("hash:    " + hash + " " + userData.Password)
		fmt.Println("login = " + userData.Login)
		fmt.Errorf("file: SqlConnection, failed to get user: %w", err)
		return structs.User{}, err
	}

	if errCompareHash := bcrypt.CompareHashAndPassword([]byte(hash), []byte(userData.Password)); errCompareHash != nil {
		if errors.Is(errCompareHash, bcrypt.ErrMismatchedHashAndPassword){
			return structs.User{}, ErrCompareHashAndPassword
		}
		return structs.User{}, errCompareHash
	}

	return structs.User{
		Login: userData.Login,
	}, nil
}

func (dataBase *DB) SelectAllUsers(command string) ([]structs.User, error) {
	var	client []structs.User

	if errSelect := dataBase.conn.Select(&client, command); errSelect != nil {
		fmt.Println("Select command failed: $1}\n", errSelect)
		return nil, errSelect
	}

	return client, nil
}

func (dataBase *DB) SelectUserByID(id string) (structs.User, error){
	var client structs.User

	if errGet := dataBase.conn.Get(&client, `SELECT id, login, password FROM users WHERE id = $1`, id);
	errGet != nil{
		fmt.Errorf("daild get user by id: %w\n", errGet)
		return client, errGet
	}

	return client, nil
}

func SqlPGX() {
	ctxwt, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connUri := "postgresql://habitov:habitov@95.217.232.188:7777/habitov"

	conn, err := pgx.Connect(ctxwt, connUri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable connect database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var name string
	err = conn.QueryRow(ctxwt, "select login from usersmora where id = $1", 2).Scan(&name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	}

	fmt.Println(name)

	rows, err := conn.Query(ctxwt, `
	SELECT login, password
	FROM usersmora`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query faild: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			login    string
			password string
		)

		if err := rows.Scan(&login, &password); err != nil {
			fmt.Fprintf(os.Stderr, "Scan faild: %v\n", err)
			return
		}

		fmt.Println(login, password)
	}

	if rows.Err() != nil {
		fmt.Fprintf(os.Stderr, "Scan faild: %v\n", err)
		return
	}

}

func HashPassword(password string) (string, error){

	hash, errGenerateHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errGenerateHash != nil {
		return "", errGenerateHash
	}

	return string(hash), nil
}