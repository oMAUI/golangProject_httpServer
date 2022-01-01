package ErrorPorcessing

import (
	"dbProject/MyHttp"
	"dbProject/structs"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

func HttpError(w http.ResponseWriter, err error, msgForLogger string, msgForResponse string, code int) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("1")
	ce := structs.CustomError{
		Message: msgForResponse,
	}

	fmt.Println("2")
	res, errGetJson := MyHttp.GetJsonByte(ce)
	if errGetJson != nil {
		zap.S().Errorw("marshal", "error", errGetJson)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}

	fmt.Println("3")
	fmt.Println(msgForLogger + ": " + err.Error())
	w.WriteHeader(code)
	w.Write(res)
}
