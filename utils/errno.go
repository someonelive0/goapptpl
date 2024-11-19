package utils

import (
	"encoding/json"
	"errors"
	"time"
)

type ErrorLog struct {
	Timestamp time.Time `json:"timestamp"`
	Code      int       `json:"code"`
	Msg       string    `json:"msg"`
}

func MkErrorLog(code int, msg string) []byte {
	b, _ := json.Marshal(ErrorLog{
		Timestamp: time.Now(),
		Code:      code,
		Msg:       msg,
	})
	return b
}

func MkError(code int, msg string) error {
	b, _ := json.Marshal(ErrorLog{
		Timestamp: time.Now(),
		Code:      code,
		Msg:       msg,
	})
	return errors.New(string(b))
}
