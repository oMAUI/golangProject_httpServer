package SqlCommand

import (
	"database/sql"
	"dbProject/models/ToDoStruct"
	"dbProject/models/UserStruct"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"os"
	"time"
)

type DB struct {
	conn *sqlx.DB
}

func Connection(uri string) (*DB, error) {
	connUri := uri //"postgresql://selectel:selectel@192.168.3.30:5432/"

	dataBase, err := sqlx.Connect("pgx", connUri)
	if err != nil {
		return nil, err
	}

	dataBase.SetMaxOpenConns(1)
	dataBase.SetMaxIdleConns(3)

	//driver, errDriver := postgres.WithInstance(dataBase.DB, &postgres.Config{
	//	DatabaseName: "postgres",
	//	SchemaName: "public",
	//})
	//if errDriver != nil {
	//	return nil, fmt.Errorf("migrate instance: %w", errDriver)
	//}
	//
	//m, err := migrate.NewWithDatabaseInstance(
	//	"file://migrations",
	//	"postgres", driver)
	//if err != nil {
	//	return nil, fmt.Errorf("migrate: %w", err)
	//}
	//if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
	//	return nil, fmt.Errorf("migrate up: %w", err)
	//}

	return &DB{
		conn: dataBase,
	}, nil
}

func (dataBase *DB) CreateUser(ctx context.Context, user UserStruct.UserFromBody) (UserStruct.User, error) {
	var id string

	hash, errHash := HashPassword(user.Password)
	if errHash != nil {
		fmt.Println("failed to generate hash %w", errHash.Error())
		return UserStruct.User{}, errHash
	}

	if errCreate := dataBase.conn.GetContext(ctx, &id,
		`INSERT INTO users(login, password)
				VALUES ($1, $2)
				RETURNING id`, user.Login, hash); errCreate != nil {
		fmt.Println("failed to create user %w", errCreate.Error())
		return UserStruct.User{}, errCreate
	}

	User := UserStruct.User{
		ID:       id,
		Login:    user.Login,
		Password: hash,
	}

	return User, nil
}

func (dataBase *DB) AuthorizationUser(ctx context.Context, userData UserStruct.UserFromBody) (UserStruct.User, error) {
	var gettingUser UserStruct.User

	if err := dataBase.conn.GetContext(ctx, &gettingUser, "SELECT id, login, password FROM users WHERE login = $1",
		userData.Login); err != nil {
		//fmt.Println("hash:    " + hash + " " + userData.Password)
		fmt.Println("login = " + userData.Login)
		fmt.Errorf("file: SqlConnection, failed to get user: %w", err)
		return UserStruct.User{}, err
	}

	if errCompareHash := bcrypt.CompareHashAndPassword([]byte(gettingUser.Password), []byte(userData.Password)); errCompareHash != nil {
		if errors.Is(errCompareHash, bcrypt.ErrMismatchedHashAndPassword) {
			return UserStruct.User{}, fmt.Errorf("wrong password")
		}
		return UserStruct.User{}, errCompareHash
	}

	return UserStruct.User{
		ID:       gettingUser.ID,
		Login:    userData.Login,
		Password: gettingUser.Password,
	}, nil
}

func (dataBase *DB) InTx(ctx context.Context, isolation sql.IsolationLevel, f func(tx sqlx.Tx) error) error {
	tx, err := dataBase.conn.BeginTxx(ctx, &sql.TxOptions{
		Isolation: isolation,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := f(*tx); err != nil {
		if errRoll := tx.Rollback(); errRoll != nil {
			return fmt.Errorf("rollback tx: %v (error: %w)", errRoll, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (dataBase *DB) CreateToDoList(ctx context.Context, todo ToDoStruct.ToDoList, UserRequest UserStruct.User) (ToDoStruct.ToDoList, error) {
	var todoList ToDoStruct.ToDoList

	if err := dataBase.InTx(ctx, sql.LevelReadCommitted, func(tx sqlx.Tx) error {
		if err := tx.Get(&todoList, `
		INSERT INTO todo_lists (title, created_by)
		VALUES ($1, $2)
		RETURNING id, title, created_at, created_by`, todo.Title, UserRequest.ID); err != nil {
			return fmt.Errorf("insert todo_lists: %w", err)
		}

		if _, err := tx.Exec(`INSERT INTO users_rights 
		(users_id, todo_lists_id, rights)
		VALUES ($1, $2, $3)`, UserRequest.ID, todoList.ID, UserStruct.Owner); err != nil {
			return fmt.Errorf("insert user_rights: %w", err)
		}
		return nil
	}); err != nil {
		return ToDoStruct.ToDoList{}, err
	}

	return todoList, nil
}

func (dataBase *DB) GetRights(ctx context.Context, ToDoListID, UserID string) (UserStruct.Rights, error) {
	var right UserStruct.Rights
	if errGetRights := dataBase.conn.GetContext(ctx, &right, `
		SELECT rights
		FROM users_rights
		WHERE todo_lists_id = $1 AND users_id = $2
	`, ToDoListID, UserID); errGetRights != nil {
		if errors.Is(errGetRights, sql.ErrNoRows) {
			return UserStruct.NoRights, nil
		}

		return UserStruct.NoRights, fmt.Errorf("get user rights: %w", errGetRights)
	}

	return right, nil
}

func (dataBase *DB) AvailableToDoList(ctx context.Context, UserID string) ([]ToDoStruct.ToDoList, error) {
	var tl []ToDoStruct.ToDoList
	if errGetTL := dataBase.conn.SelectContext(ctx, &tl, `
		SELECT tl.id, tl.title, tl.created_at, tl.created_by
		FROM todo_lists tl
		INNER JOIN users_rights ur ON  tl.created_by = ur.users_id 
			AND tl.id = ur.todo_lists_id
		WHERE tl.created_by = $1
	`, UserID); errGetTL != nil {
		return nil, fmt.Errorf("get available todo list: %w", errGetTL)
	}

	return tl, nil
}

func (dataBase *DB) GetToDoListToDo(ctx context.Context, todoListID string) ([]ToDoStruct.ToDo, error) {
	var td []ToDoStruct.ToDo
	if err := dataBase.conn.SelectContext(ctx, &td, `
	SELECT id, title, description, checked, 
		todo_lists_id, created_at, created_by, 
		updated_at, updated_by
	FROM todos
	WHERE todo_lists_id = $1
	`, todoListID); err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}

	return td, nil
}

func (dataBase *DB) CreateRights(ctx context.Context, rights ToDoStruct.UserRights) error {
	if _, err := dataBase.conn.ExecContext(ctx, `
	INSERT INTO users_rights (users_id, todo_lists_id, rights)
	VALUES ($1, $2, $3)`,
		rights.UserID, rights.TODOListID, rights.Rights); err != nil {
		return fmt.Errorf("create user rights: %w", err)
	}

	return nil
}

func (dataBase *DB) SelectAllUsers(command string) ([]UserStruct.User, error) {
	var client []UserStruct.User

	if errSelect := dataBase.conn.Select(&client, command); errSelect != nil {
		fmt.Println("Select command failed: $1}\n", errSelect)
		return nil, errSelect
	}

	return client, nil
}

func (dataBase *DB) SelectUserByID(id string) (UserStruct.User, error) {
	var client UserStruct.User

	if errGet := dataBase.conn.Get(&client, `SELECT id, login, password FROM users WHERE id = $1`, id); errGet != nil {
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

func HashPassword(password string) (string, error) {

	hash, errGenerateHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errGenerateHash != nil {
		return "", errGenerateHash
	}

	return string(hash), nil
}
