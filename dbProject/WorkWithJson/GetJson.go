package WorkWithJson

import "encoding/json"

func GetJsonByte(v interface{}) ([]byte, error) {
	usersJson, errJson := json.Marshal(v)
	if errJson != nil {
		return nil, errJson
	}

	return usersJson, nil
}
