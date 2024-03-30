package kvstore

import (
	"reflect"

	"github.com/necrobits/x/errors"
)

const (
	EPointerExpected = "pointer_expected"
	EUnassignable    = "unassignable"
)

func Copy(src any, dest any) error {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest)

	if destVal.Kind() != reflect.Ptr {
		return errors.B().
			Code(EPointerExpected).
			Msg("destination should be a pointer").Build()
	}

	// Check if src is assignable to dest
	if !srcVal.Type().AssignableTo(destVal.Elem().Type()) {
		return errors.B().
			Code(EUnassignable).
			Msg("source is not assignable to destination").Build()
	}

	destVal.Elem().Set(srcVal)

	return nil
}
