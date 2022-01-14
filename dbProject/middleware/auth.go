package auth

import (
	"context"
	"dbProject/models/UserStruct"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"strings"
)

func Auth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			bearer := request.Header.Get("Authorization")
			s := strings.Split(bearer, " ")

			if len(s) != 2 {
				log.Default().Println("Split")
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			token, errParseToken := jwt.ParseWithClaims(s[1], &UserStruct.WithClaims{}, func(t *jwt.Token) (interface{}, error){
				return UserStruct.AuthToken, nil
			})
			if errParseToken != nil {
				log.Default().Println("token parse failed: ", errParseToken)
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				log.Default().Println("not valid token")
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*UserStruct.WithClaims)
			if !ok {
				log.Default().Println("claims: ", ok)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Default().Println("claims: ", claims)

			request = request.WithContext(context.WithValue(request.Context(), UserStruct.CtxKey(), claims.ToUser()))
			next.ServeHTTP(writer, request)
		})
	}
}
