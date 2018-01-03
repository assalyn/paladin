package paladin

import (
	"fmt"
	"reflect"
)

func showStruct(value reflect.Value) {
	fmt.Printf("[showStruct]")
	for i := 0; i < value.NumField(); i++ {
		fmt.Printf("%#v ", value.Field(i).Interface())
	}
	fmt.Printf("\n")
}

func showSlice(value reflect.Value) {
	fmt.Printf("[showSlice]")
	for i := 0; i < value.Len(); i++ {
		fmt.Printf("%#v ", value.Index(i).Interface())
	}
	fmt.Printf("\n")
}
