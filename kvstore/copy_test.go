package kvstore

import (
	"encoding/json"
	"testing"

	"github.com/necrobits/x/errors"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("NotPointer", func(t *testing.T) {
		var src string = "test"
		var dst string
		err := Copy(src, dst)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, EPointerExpected) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("Unassignable_Simple", func(t *testing.T) {
		var src string = "test"
		var dst int
		err := Copy(src, &dst)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, EUnassignable) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("Success_String", func(t *testing.T) {
		var src string = "test"
		var dst string
		err := Copy(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if src != dst {
			t.Errorf("expected src and dst to be equal")
		}
		require.Equal(t, src, dst)
	})

	t.Run("Success_Interface", func(t *testing.T) {
		var src interface{} = "test"
		var dst interface{}
		err := Copy(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if src != dst {
			t.Errorf("expected src and dst to be equal")
		}
		require.Equal(t, src, dst)
	})

	t.Run("Success_Map", func(t *testing.T) {
		var src map[string]any = map[string]any{"test": "test"}
		var dst map[string]any
		err := Copy(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if src["test"] != dst["test"] {
			t.Errorf("expected src and dst to be equal")
		}
		require.Equal(t, src, dst)
	})

	t.Run("Success_Complex", func(t *testing.T) {
		type TestStruct struct {
			StrField  string
			JsonField json.RawMessage
		}
		type JsonStruct struct {
			Field string
		}
		jsonField := JsonStruct{
			Field: "test",
		}
		jsonFieldBytes, err := json.Marshal(jsonField)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		var src any = TestStruct{
			StrField:  "test",
			JsonField: jsonFieldBytes,
		}
		var dst TestStruct
		err = Copy(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		require.Equal(t, src, dst)
	})
}
