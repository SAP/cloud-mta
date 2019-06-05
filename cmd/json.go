package commands

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/DisposaBoy/JsonConfigReader"
)

func ParseJSONWithComments(input []byte, obj interface{}) error {
	r := JsonConfigReader.New(bytes.NewReader(input))
	return json.NewDecoder(r).Decode(obj)
}

func ConvertToJsonSafe(val interface{}) interface{} {
	switch v := val.(type) {
	case map[interface{}]interface{}:
		//valMap :=
		res := map[string]interface{}{}
		for k, v := range v {
			switch v2 := v.(type) {
			case map[interface{}]interface{}, []interface{}:
				res[fmt.Sprint(k)] = ConvertToJsonSafe(v2)
			default:
				res[fmt.Sprint(k)] = v
			}
		}
		return res
	case []interface{}:
		for k, v2 := range v {
			v[k] = ConvertToJsonSafe(v2)
		}
		return v
	}
	return val
}
