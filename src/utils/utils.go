package utils

import "encoding/json"

type errorMsg struct {
	Error string `json:"error"`
}

func NewWrappedError(message string) []byte {
	res, _ := json.Marshal(errorMsg{Error: message})
	return res
}

func WrapError(err error) []byte {
	return NewWrappedError(err.Error())
}
