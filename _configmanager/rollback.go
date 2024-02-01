package old_configmanager

import "reflect"

type Rollback struct {
	value    reflect.Value
	oldValue reflect.Value
	// key is only valid if field is a map
	key           reflect.Value
	sliceAppended bool
}

type RollbackList []Rollback

func (l RollbackList) rollback() {
	for i := len(l) - 1; i >= 0; i-- {
		rb := l[i]
		if rb.key.IsValid() {
			rb.value.SetMapIndex(rb.key, rb.oldValue)
		} else if rb.sliceAppended {
			rb.value.SetLen(rb.value.Len() - 1)
		} else {
			rb.value.Set(rb.oldValue)
		}
	}
}
