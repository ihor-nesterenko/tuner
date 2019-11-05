package tuner

import "reflect"

func isPointer(i interface{}) bool {
	return reflect.ValueOf(i).Kind() == reflect.Ptr
}
