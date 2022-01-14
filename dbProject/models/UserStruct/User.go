package UserStruct

import (
	"context"
	"dbProject/WorkWithJson"
	"github.com/golang-jwt/jwt"
	"time"
)

type User struct {
	ID        string             `db:"id" json:"id"`
	Login     string             `db:"login" json:"login"`
	Password  string             `db:"password" json:"password"`
}

type UserFromBody struct {
	Login    string `db:"login" json:"login"`
	Password string `db:"password" json:"password"`
}

type TokenResp struct {
	Token string `json:"token"`
}

type Rights int

var (
	NoRights   Rights = -1
	Read       Rights = 0
	Write      Rights = 1
	AdminRead  Rights = 2
	AdminWrite Rights = 3
	Owner      Rights = 4

	AuthToken = []byte("maui")
)

type WithClaims struct {
	jwt.StandardClaims
	ID    string `json:"id"`
	Login string `json:"login"`
}

func (w *WithClaims) ToUser() User {
	return User{
		ID:    w.ID,
		Login: w.Login,
	}
}

func CheckRights(Need Rights, Has Rights) bool {
	return Has >= Need
}

func CanGiveRights(with Rights, give Rights) bool {
	if with == Owner {
		return true
	}

	if with == AdminWrite && (give == Read || give == Write){
		return true
	}

	if with == AdminRead && give == Read {
		return true
	}

	return false
}

func (u User) GetToken() ([]byte, error) {
	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, WithClaims{
		ID:    u.ID,
		Login: u.Login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	})

	token, errSigningToken := tokenWithClaims.SignedString(AuthToken)
	if errSigningToken != nil {
		//HttpError(w, errSigningToken, "failed to Signing String", "", http.StatusInternalServerError)
		return nil, errSigningToken
	}

	tokenResp := TokenResp{
		Token: token,
	}

	tokenJson, errJson := WorkWithJson.GetJsonByte(tokenResp)
	if errJson != nil {
		return nil, errJson
	}

	return tokenJson, nil
}

func FromCtx(ctx context.Context) User {
	return ctx.Value(CtxKey()).(User)
}

type Key struct {
	K string
}

func CtxKey() Key {
	return Key{K: "id"}
}
