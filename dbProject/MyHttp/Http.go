package MyHttp

import (
	"dbProject/ErrorPorcessing"
	"dbProject/WorkWithJson"
	"dbProject/WorkWithToken"
	"dbProject/structs"
	"encoding/json"
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
	SelectAllUsers(string)([]structs.User, error)
	SelectUserByID(string)(structs.User, error)
	CreateUser(user structs.UserFromBody) (structs.User, error)
	AuthorizationUser(user structs.UserFromBody) (structs.User, error)
}

type Route struct{
	DB DbInterface
}

func MyRequest(ro Route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/Authorization", func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()

		var userData structs.UserFromBody
		if errUnmarshBody := UnmarshalBody(w, r.Body, &userData); errUnmarshBody != nil {
			return
		}

		User, errAuth := ro.DB.AuthorizationUser(userData)
		if errAuth != nil {
			ErrorPorcessing.HttpError(w, errAuth, "", "wrong data", http.StatusBadRequest)
			return
		}

		tokenJson, errGetToken := WorkWithToken.GetToken(User)
		if errGetToken != nil {
			ErrorPorcessing.HttpError(w, errGetToken, "failed to Signing String", "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJson)
	})

	router.Post("/signup", func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()

		var user structs.UserFromBody
		if errUnmarshUser := UnmarshalBody(w, r.Body, &user); errUnmarshUser != nil {
			return
		}

		User, errCreateUser := ro.DB.CreateUser(user)
		if errCreateUser != nil {
			ErrorPorcessing.HttpError(w, errCreateUser, errCreateUser.Error(), "server error", http.StatusInternalServerError)
			return
		}

		tokenJson, errGetToken := WorkWithToken.GetToken(User)
		if errGetToken != nil {
			ErrorPorcessing.HttpError(w, errGetToken, "failed to get token in Authorization", "Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJson)
	})

	router.Get("/users", func(w http.ResponseWriter, r *http.Request){
		var users, err = ro.DB.SelectAllUsers("SELECT * FROM users")
		if err != nil{
			fmt.Println("file: http, func MyRequest/users, error: $1", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userJson, errUsersJson := WorkWithJson.GetJsonByte(users)
		if errUsersJson != nil {
			ErrorPorcessing.HttpError(w, errUsersJson, "", "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(userJson)
	})

	return router
}

func UnmarshalBody(w http.ResponseWriter,r io.Reader, v interface{}) error {
	resp, errResp := ioutil.ReadAll(r)
	if errResp != nil {
		ErrorPorcessing.HttpError(w, errResp, "failed to get body", "Bad Request", http.StatusBadRequest)
		return errResp
	}

	if errUnmarshalJson := json.Unmarshal(resp, v); errUnmarshalJson != nil {
		ErrorPorcessing.HttpError(w, errUnmarshalJson, "failed to get Json in Authorization", "Server Error", http.StatusInternalServerError)
		return errUnmarshalJson
	}

	return nil
}