package MyHttp

import (
	"context"
	"database/sql"
	"dbProject/ErrorPorcessing"
	auth "dbProject/middleware"
	"dbProject/models/ToDoStruct"
	"dbProject/models/UserStruct"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jmoiron/sqlx"
)

//TODO: 1) рефакторинг кода; вынести повторяющийся код

type DbInterface interface {
	//SelectAllUsers(context.Context, string) ([]structs.User, error)
	//SelectUserByID(context.Context, string) (structs.User, error)
	CreateUser(context.Context, UserStruct.UserFromBody) (UserStruct.User, error)
	AuthorizationUser(context.Context, UserStruct.UserFromBody) (UserStruct.User, error)
	CreateToDoList(context.Context, ToDoStruct.ToDoList, UserStruct.User) (ToDoStruct.ToDoList, error)
	GetRights(context.Context, string, string) (UserStruct.Rights, error)
	AvailableToDoList(context.Context, string) ([]ToDoStruct.ToDoList, error)
	GetToDoListToDo ( context.Context, string) ([]ToDoStruct.ToDo, error)
	CreateRights(context.Context,  ToDoStruct.UserRights) error
}

type Route struct{
	DB DbInterface
}

func MyRequest(ro Route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Group(func(route chi.Router){
		route.Use(auth.Auth())

		route.Post("/todo_list", func (w http.ResponseWriter, r *http.Request){
			defer r.Body.Close()

			var toDoList ToDoStruct.ToDoList
			if errUnmarshal := UnmarshalBody(r.Body, &toDoList); errUnmarshal != nil {
				ErrorPorcessing.HttpError(w, errUnmarshal, "failed to unmarshal todolist",
					"server error", http.StatusInternalServerError)
				return
			}

			userCtx := UserStruct.FromCtx(r.Context())
			todoList, errCreateToDoList := ro.DB.CreateToDoList(r.Context(), toDoList,
				UserStruct.User{
					ID: userCtx.ID,
					Login: userCtx.Login,
			})
			if errCreateToDoList != nil {
				ErrorPorcessing.HttpError(w, errCreateToDoList, "failed to create to do list", "failed to create to do list", http.StatusInternalServerError)
				return
			}

			respJson, errGetJson := json.Marshal(todoList)
			if errGetJson != nil {
				ErrorPorcessing.HttpError(w, errGetJson, "failed to get json in create todolist",
					"server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application-json")
			w.Write(respJson)
		})

		route.Post("/user_rights", func(w http.ResponseWriter, r *http.Request){
			defer r.Body.Close()

			var userRights ToDoStruct.UserRights
			if errUnmarshalBody := UnmarshalBody(r.Body, &userRights); errUnmarshalBody != nil {
				ErrorPorcessing.HttpError(w, errUnmarshalBody, "unmarshal body in user_rights",
					"server error", http.StatusInternalServerError)
				return
			}

			userCtx := UserStruct.FromCtx(r.Context())
			rights, errGetRights := ro.DB.GetRights(r.Context(), userRights.TODOListID, userCtx.ID)
			if errGetRights != nil {
				ErrorPorcessing.HttpError(w, errGetRights, "get rights",
					"wrong data", http.StatusBadRequest)
				return
			}

			if ok := UserStruct.CanGiveRights(rights, userRights.Rights); !ok {
				ErrorPorcessing.HttpError(w, fmt.Errorf("no rights"), "unmarshal body in user_rights",
					"no rights", http.StatusForbidden)
				return
			}

			if errCreateRights := ro.DB.CreateRights(r.Context(), userRights); errCreateRights != nil {
				ErrorPorcessing.HttpError(w, errCreateRights, "create rights",
					"server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)

			if errCloseBody := r.Body.Close(); errCloseBody != nil {
				ErrorPorcessing.HttpError(w, errCloseBody, "close body",
					"server error", http.StatusInternalServerError)
				return
			}
		})

		//Get available todo_lists
		route.Get("/todo_list", func(w http.ResponseWriter, r *http.Request){
			userCtx := UserStruct.FromCtx(r.Context())
			todoLists, errGetTodoLists := ro.DB.AvailableToDoList(r.Context(), userCtx.ID)
			if errGetTodoLists != nil {
				ErrorPorcessing.HttpError(w, errGetTodoLists, "get available todo_lists",
					"wrong data", http.StatusBadRequest)
				return
			}

			resp, errResp := json.Marshal(todoLists)
			if errResp != nil {
				ErrorPorcessing.HttpError(w, errResp, "marshal available todo list",
					"server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application-json")
			w.Write(resp)
		})

		route.Get("/todo_list/{id}", func(w http.ResponseWriter, r *http.Request){
			tlID := chi.URLParam(r, "id")

			userCtx := UserStruct.FromCtx(r.Context())
			right, errGetRight := ro.DB.GetRights(r.Context(), tlID, userCtx.ID)
			if errGetRight != nil {
				if errors.Is(errGetRight, sql.ErrNoRows){
					ErrorPorcessing.HttpError(w, errGetRight, "get rights", "no rights",
						http.StatusBadRequest)
					return
				}

				ErrorPorcessing.HttpError(w, errGetRight, "get rights",
					"server error", http.StatusInternalServerError)
				return
			}

			if ok := UserStruct.CheckRights(UserStruct.Read, right); !ok {
				ErrorPorcessing.HttpError(w, fmt.Errorf("access forbidden"), "check rights", "not forbidden",
					http.StatusForbidden)
				return
			}

			tl, errGetTl := ro.DB.GetToDoListToDo(r.Context(), tlID)
			if errGetTl != nil {
				ErrorPorcessing.HttpError(w, errGetTl, "get to do", "server error",
					http.StatusInternalServerError)
				return
			}

			resp, errReq := json.Marshal(tl)
			if errReq != nil {
				ErrorPorcessing.HttpError(w, errReq, "marshal to do", "server error",
					http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application-json")
			w.Write(resp)
		})
	})

	router.Get("/Authorization", func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()

		var userData UserStruct.UserFromBody
		if errUnmarshBody := UnmarshalBody(r.Body, &userData); errUnmarshBody != nil {
			return
		}

		User, errAuth := ro.DB.AuthorizationUser(r.Context(), userData)
		if errAuth != nil {
			ErrorPorcessing.HttpError(w, errAuth, "", "wrong data", http.StatusBadRequest)
			return
		}

		tokenJson, errGetToken := User.GetToken()
		if errGetToken != nil {
			ErrorPorcessing.HttpError(w, errGetToken, "failed to Signing String", "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJson)
	})

	router.Post("/signup", func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()

		var user UserStruct.UserFromBody
		if errUnmarshalUser := UnmarshalBody(r.Body, &user); errUnmarshalUser != nil {
			return
		}

		User, errCreateUser := ro.DB.CreateUser(r.Context(), user)
		if errCreateUser != nil {
			ErrorPorcessing.HttpError(w, errCreateUser, errCreateUser.Error(), "server error", http.StatusInternalServerError)
			return
		}

		tokenJson, errGetToken := User.GetToken()
		if errGetToken != nil {
			ErrorPorcessing.HttpError(w, errGetToken, "failed to get token in Authorization", "Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJson)
	})

	return router
}

func UnmarshalBody(r io.Reader, v interface{}) error {
	resp, errResp := ioutil.ReadAll(r)
	if errResp != nil {
		//ErrorPorcessing.HttpError(w, errResp, "failed to get body", "Bad Request", http.StatusBadRequest)
		return fmt.Errorf("server error: %w", errResp)
	}

	if errUnmarshalJson := json.Unmarshal(resp, v); errUnmarshalJson != nil {
		//ErrorPorcessing.HttpError(w, errUnmarshalJson, "failed to get Json in Authorization", "Server Error", http.StatusInternalServerError)
		return fmt.Errorf("server error: %w", errUnmarshalJson)
	}

	return nil
}