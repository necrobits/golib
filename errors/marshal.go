package errors

import (
	"encoding/json"
	"errors"
)

type serializedError struct {
	Op   string      `json:"op,omitempty"`
	Code string      `json:"code,omitempty"`
	Msg  string      `json:"msg,omitempty"`
	Err  interface{} `json:"error,omitempty"`
}

func (e *appError) MarshalJSON() ([]byte, error) {
	t := serializedError{
		Op:   e.Op,
		Code: e.Code,
		Msg:  e.Msg,
	}
	if e.Err != nil {
		if _, ok := e.Err.(*appError); ok {
			_, err := e.Err.(*appError).MarshalJSON()
			if err != nil {
				return nil, err
			}
			t.Err = e.Err
		} else {
			t.Err = e.Err.Error()
		}
	}
	return json.Marshal(t)
}

func (e *appError) UnmarshalJSON(data []byte) error {
	t := new(serializedError)
	err := json.Unmarshal(data, t)
	if err != nil {
		return err
	}
	e.Op = t.Op
	e.Code = t.Code
	e.Msg = t.Msg
	e.Err = decodeErr(t.Err)
	return nil
}

func decodeErr(err interface{}) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(string); ok {
		return errors.New(e)
	}
	if errMap, ok := err.(map[string]interface{}); ok {
		e := new(appError)
		if op, ok := errMap["op"].(string); ok {
			e.Op = op
		}
		if code, ok := errMap["code"].(string); ok {
			e.Code = code
		}
		if msg, ok := errMap["msg"].(string); ok {
			e.Msg = msg
		}
		e.Err = decodeErr(errMap["error"])
	}
	return nil
}
