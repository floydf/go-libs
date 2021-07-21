package lib

import (
	"encoding/json"
)

func Jsonify(o interface{}) string {
	b, _ := json.MarshalIndent(o, "", "    ")
	return string(b)
}
