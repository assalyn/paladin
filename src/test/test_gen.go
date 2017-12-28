package main

import (
	"reflect"

	"frm/plog"

	"fmt"

	"github.com/dave/jennifer/jen"
)

// 定义类型
// 根据类型的反射定义来生成结构定义文件

type BBB struct {
	Data int
}

type AAA struct {
	Int    int
	String string
	Bbb    BBB
}

func main() {
	plog.InitLog("app.log", plog.LOG_TRACE)

	c := NewCoder("equip")
	c.GenStruct(AAA{})
	c.Output(false)
}

type Coder struct {
	jfile    *jen.File
	fileName string
}

func NewCoder(fileName string) *Coder {
	c := new(Coder)
	c.fileName = fileName + ".dbc.go"
	c.jfile = jen.NewFile("dbc")
	return c
}

func (p *Coder) GenStruct(obj interface{}) {
	t := reflect.TypeOf(obj)
	//v := reflect.ValueOf(obj)
	fmt.Println("struct name", t.Name())

	fields := make([]jen.Code, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fmt.Println("field", i, " ", t.Field(i).Name, t.Field(i).Type.Kind().String())
		switch t.Field(i).Type.Kind().String() {
		case "int":
			fields[i] = jen.Id(t.Field(i).Name).Int()

		case "string":
			fields[i] = jen.Id(t.Field(i).Name).String()

		case "struct":
			fmt.Println(t.Field(i).Name, t.Field(i).Type.String())
			//for j := 0; j < elem.NumField(); j++ {
			//	fmt.Println(elem.Field(j).Type())
			//}
			//for j := 0; j < subType.NumField(); j++ {
			//	fmt.Println("subfield", j, " ", t.Field(j).Name, t.Field(j).Type)
			//}
			fields[i] = jen.Id(t.Field(i).Name).Qual("", "BBB")
			// 记录一下，结束后要生成这个数据结构的

		default:
			plog.Panic("not support type", t.Field(i).Type)
		}
	}
	p.jfile.Type().Id(t.Name()).Struct(fields...)

}

func (p *Coder) Output(writeFile bool) {
	if writeFile {
		if err := p.jfile.Save(p.fileName); err != nil {
			plog.Error(err)
		}
	} else {
		fmt.Printf("\n\n%#v", p.jfile)
	}
}
