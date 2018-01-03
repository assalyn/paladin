package main

import (
	"cmn"
	"encoding/json"
	"fmt"
	"frm/plog"
	"reflect"
	"regexp"
)

type Point struct {
	X int
	Y int
}

type XXX struct {
	Value int
	Point Point
}

func main() {
	plog.InitLog("test.log", plog.LOG_TRACE)

	fmt.Println(regexp.Match("(?i:^\\[Rate\\])", []byte("[rate]#1")))
	/*fields := make([]reflect.StructField, 2)
	fields[0].Name = "AAA"
	fields[0].Type = reflect.TypeOf(int(0))
	fields[1].Name = "BBB"
	fields[1].Type = reflect.TypeOf("")
	t := reflect.StructOf(fields)*/

	//coder := paladin.NewCodeBuilder("aaa")
	//coder.DebugType(reflect.TypeOf(&XXX{}), "Aaa")
}

func TestCamelName() {
	fmt.Println(cmn.CamelName("hello"))
	fmt.Println(cmn.CamelName("hello_world"))
	fmt.Println(cmn.CamelName("hello world"))
	fmt.Println(cmn.CamelName("a_b_"))
}

type Test struct {
	S string
	X byte
	Y uint64
}

func test() {
	fields := []reflect.StructField{
		{
			Name: "S",
			Tag:  "s",
			Type: reflect.TypeOf(""),
		},
		{
			Name: "X",
			Tag:  "x",
			Type: reflect.TypeOf(byte(0)),
		},
		{
			Name: "Y",
			Type: reflect.TypeOf(uint64(0)),
		},
		{
			Name: "Z",
			Type: reflect.TypeOf([3]uint16{}),
		},
	}
	typ := reflect.StructOf(fields)
	fmt.Println("typ.string", typ.String(), "typ.name", typ.Name(), "typ.PkgPath", typ.PkgPath())
	fmt.Printf("%v\n", typ)

	value := reflect.New(typ) // 一个实例的value
	elem := value.Elem()
	fmt.Println(elem)
	elem.Field(0).SetString("test")
	elem.Field(1).SetUint(1)
	elem.Field(2).SetUint(2)
	sliceValue := elem.Field(3)
	sliceValue.Index(0).SetUint(1)
	sliceValue.Index(1).SetUint(2)
	sliceValue.Index(2).SetUint(3)
	marshal(value.Interface())
}

func marshal(value interface{}) {
	fmt.Printf("marshal %+v\n", value)
	rawData, err := json.Marshal(value)
	if err != nil {
		plog.Panic(err)
	}
	fmt.Println(string(rawData))
}
