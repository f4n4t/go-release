package utils

import (
	"encoding/json"
	"fmt"
)

// GetJSONString returns a JSON string representation of an object.
func GetJSONString(obj any, withIndent bool) (string, error) {
	var (
		out []byte
		err error
	)

	switch {
	case withIndent:
		out, err = json.MarshalIndent(obj, "", "\t")

	default:
		out, err = json.Marshal(obj)
	}
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}

	return string(out), nil
}
