package utils

import (
	"fmt"
	"reflect"
)

func EnsureIsSliceOrArray(obj interface{}) (res reflect.Value) {
	vo := reflect.ValueOf(obj)
	for vo.Kind() == reflect.Ptr || vo.Kind() == reflect.Interface {
		vo = vo.Elem()
	}
	k := vo.Kind()
	if k != reflect.Slice && k != reflect.Array {
		panic(fmt.Sprintf("obj required slice or array type, but got %v", vo.Type()))
	}
	res = vo
	return
}
