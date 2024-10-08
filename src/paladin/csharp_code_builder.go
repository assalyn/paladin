package paladin

import (
	"assalyn/paladin/cmn"
	"assalyn/paladin/frm/plog"
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"
)

type CsharpCodeBuilder struct {
	file       *CodeWriter
	codeDir    string
	dataDir    string
	fileName   string
	structName string // 给定的外层结构名, 为解决创建的匿名结构问题
	ns         string // namespace
}

func NewCsharpCodeBuilder(codeDir string, dataDir string, fileName string, ns string) *CsharpCodeBuilder {
	c := new(CsharpCodeBuilder)
	c.codeDir = codeDir
	c.dataDir = dataDir
	c.fileName = fileName
	c.ns = ns
	c.file = NewCodeWriter()
	c.file.HeadComment("// Code generated by paladin. DO NOT EDIT.\n")
	return c
}

func (p *CsharpCodeBuilder) GenStruct(obj interface{}) {
	p.structName = reflect.TypeOf(obj).Name()
	p.GenStructWithName(obj, p.structName)
}

func (p *CsharpCodeBuilder) GenStructWithName(obj interface{}, structName string) {
	p.file.Using("using System;\n")
	p.file.Using("using System.Collections.Generic;\n")
	p.file.Namespace(p.ns)

	p.structName = cmn.CamelName(structName)
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	p.genType(t, p.structName, "")
}

func (p *CsharpCodeBuilder) GenType(t reflect.Type, structName string) {
	p.structName = structName
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	p.genType(t, p.structName, "")
}

// 这个代码问题还挺多的
func (p *CsharpCodeBuilder) genType(t reflect.Type, structName string, printPrefix string) {
	//fmt.Printf("%s[gen type %s]\n", printPrefix, t.Name())
	s := p.file.Struct(structName)
	for i := 0; i < t.NumField(); i++ {
		subField := t.Field(i)
		//fmt.Printf("%sfield %d %s %s\n", printPrefix+"  ", i, t.Field(i).Name, t.Field(i).Type.Kind().String())
		subFieldName := cmn.FirstCharLower(subField.Name)
		switch subField.Type.Kind() {
		case reflect.Struct:
			subStruct := structName + subField.Name // 这里不能用小写
			p.genType(subField.Type, subStruct, printPrefix+"  ")
			s.AddField(subStruct, subFieldName)

		case reflect.Map:
			elem := subField.Type.Elem()
			if elem.Kind() == reflect.Struct {
				mapSubStruct := structName + subField.Name
				p.genType(elem, mapSubStruct, printPrefix+"  ")
				s.AddMap(p.TypeName(subField.Type.Key()), mapSubStruct, subFieldName)
			} else {
				// 常规类型 int, float, string等
				s.AddMap(p.TypeName(subField.Type.Key()), p.TypeName(elem), subFieldName)
			}

		case reflect.Slice:
			elem := subField.Type.Elem()
			if elem.Kind() == reflect.Struct {
				sliceSubStruct := structName + subField.Name
				p.genType(elem, sliceSubStruct, printPrefix+"  ")
				s.AddSlice(sliceSubStruct, subFieldName)
			} else {
				// 常规类型 int, float, string等
				s.AddSlice(p.TypeName(elem), subFieldName)
			}

		default:
			// 基础类型
			s.AddField(p.TypeName(subField.Type), subFieldName)
		}
	}
	if structName == "" {
		structName = t.Name()
	}
	s.Save()
}

// 输出
func (p *CsharpCodeBuilder) Output() {
	if err := p.file.Save(p.codeDir + "/" + p.fileName + ".cs"); err != nil {
		plog.Error(err)
	}
}

func (p *CsharpCodeBuilder) DebugType(t reflect.Type, structName string) {
	p.structName = structName
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	p.genType(t, p.structName, "")
	fmt.Printf("%#v\n", p.file)
}

// 反射类型转换为jennifer语句
func (p *CsharpCodeBuilder) TypeToJenStatement(t reflect.Type) *jen.Statement {
	switch t.Kind() {
	case reflect.Bool:
		return jen.Bool()

	case reflect.Int,
		reflect.Int64:
		return jen.Int64()

	case reflect.Int32:
		return jen.Int()

	case reflect.String:
		return jen.String()

	case reflect.Float32:
		return jen.Float32()

	case reflect.Float64:
		return jen.Float64()

	default:
		return jen.String()
	}
}

func (p *CsharpCodeBuilder) TypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"

	case reflect.Float32:
		return "float"

	case reflect.Float64:
		return "double"

	case reflect.Int,
		reflect.Int64:
		return "long"

	case reflect.Int32:
		return "int"

	case reflect.String:
		return "string"

	default:
		plog.Panic("not support type", t)
		return "null"
	}
}

func (p *CsharpCodeBuilder) structNameToValueName(structName string) (valueName string) {
	if structName == "" {
		return ""
	}
	bs := make([]byte, 0, len(structName))
	bs = append(bs, structName[0]+32)
	bs = append(bs, structName[1:]...)
	return string(bs)
}
