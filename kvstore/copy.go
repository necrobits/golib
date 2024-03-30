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

	destVal = destVal.Elem()
	srcType := srcVal.Type()
	destType := destVal.Type()

	if srcType.AssignableTo(destType) {
		destVal.Set(srcVal)
		return nil
	} else if srcType.ConvertibleTo(destType) {
		destVal.Set(srcVal.Convert(destType))
		return nil
	}

	return errors.B().
		Code(EUnassignable).
		Msg("source and destination are not assignable").Build()
}
