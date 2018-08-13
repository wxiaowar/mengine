// +build !jsoniter

package json

import "encoding/json"

var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewEncoder    = json.NewEncoder
	NewDecoder    = json.NewDecoder
)
