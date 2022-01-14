package ToDoStruct

import (
	"dbProject/models/UserStruct"
	"time"
)

type ToDo struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	TodoListsID string    `json:"todo_lists_id" db:"todo_lists_id"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	Checked     bool      `json:"checked" db:"checked"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   time.Time `json:"deleted_by" db:"deleted_by"`
}

type ToDoList struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserRights struct {
	UserID     string            `json:"user_id" db:"user_id"`
	TODOListID string            `json:"todo_list_id" db:"todo_list_id"`
	Rights     UserStruct.Rights `json:"rights" db:"rights"`
}
