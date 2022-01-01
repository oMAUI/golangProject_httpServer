package WorkWithToken

import (
	"dbProject/WorkWithJson"
	"dbProject/structs"
	"github.com/golang-jwt/jwt"
)

func GetToken(User structs.User) ([]byte, error){
	tokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"login": User.Login,
	})

	token, errSigningToken := tokenWithClaims.SigningString()
	if errSigningToken != nil {
		//HttpError(w, errSigningToken, "failed to Signing String", "", http.StatusInternalServerError)
		return nil, errSigningToken
	}

	tokenResp := structs.TokenResp{
		Token: token,
	}

	tokenJson, errJson := WorkWithJson.GetJsonByte(tokenResp)
	if errJson != nil {
		return nil, errJson
	}

	return tokenJson, nil
}
