package SqlCommand

import (
	"context"
	"dbProject/UserStruct"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestDB_CreateToDoList(t *testing.T) {
	connUri := "postgresql://maui:maui@192.168.0.26:5432/postgres"

	dataBase, _ := sqlx.Connect("pgx", connUri)

	type fields struct {
		conn *sqlx.DB
	}
	type args struct {
		ctx         context.Context
		todo        UserStruct.ToDoList
		UserRequest UserStruct.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "salim",
			fields: fields{conn: dataBase},
			args: args{ctx: context.Background(), todo: UserStruct.ToDoList{Title: "new todo"}, UserRequest: UserStruct.User{ID: "add4e97d-1f24-4ef4-b4f6-c37689640e3f"}},


		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataBase := &DB{
				conn: tt.fields.conn,
			}
			got, err := dataBase.CreateToDoList(tt.args.ctx, tt.args.todo, tt.args.UserRequest)
			t.Errorf("got: %v", got)
			t.Errorf("err: %v", err)
		})
	}
}
