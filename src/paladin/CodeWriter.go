package paladin

import (
	"fmt"
	"assalyn/paladin/frm/plog"
	"os"
)

type CodeWriter struct {
	file         *os.File
	headComments []string
	usings       []string
	namespace    string
	contents     []string
}

func NewCodeWriter() *CodeWriter {
	w := new(CodeWriter)
	w.headComments = make([]string, 0, 4)
	w.usings = make([]string, 0, 4)
	w.contents = make([]string, 0, 12)
	return w
}

// 头注释
func (p *CodeWriter) HeadComment(comment string) *CodeWriter {
	p.headComments = append(p.headComments, comment)
	return p
}

// using package
func (p *CodeWriter) Using(pkg string) *CodeWriter {
	p.usings = append(p.usings, pkg)
	return p
}

// 设置namespace
func (p *CodeWriter) Namespace(value string) {
	p.namespace = value
}

func (p *CodeWriter) Struct(structName string) *CodeWriterStruct {
	return NewCodeWriterStruct(p, "\t", structName)
}

// 输出到具体文件
func (p *CodeWriter) Save(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		plog.Error("fail to open file!!", err)
		return err
	}
	// 输出head comment
	for _, value := range p.headComments {
		file.WriteString(value)
	}

	// 输出usings
	for _, value := range p.usings {
		file.WriteString(value)
	}
	file.WriteString("\n")

	// 输出namespace开始
	file.WriteString(fmt.Sprintf("namespace %s {\n", p.namespace))

	// 输出contents
	for _, value := range p.contents {
		file.WriteString(value)
	}

	// 输出namespace结束
	file.WriteString("}")
	return nil
}

func (p *CodeWriter) writeLine(content string) {
	p.contents = append(p.contents, content)
}

type CodeNode struct {
	fieldEnum string
	keyName   string
	typeName  string
	fieldName string
}
type CodeWriterStruct struct {
	owner *CodeWriter

	prefix     string
	structName string
	idx        int
	fields     []CodeNode // fieldName -> field type name
}

// 创建struct结构体
func NewCodeWriterStruct(owner *CodeWriter, prefix string, structName string) *CodeWriterStruct {
	s := new(CodeWriterStruct)
	s.prefix = prefix
	s.owner = owner
	s.structName = structName
	s.fields = make([]CodeNode, 0, 10)
	return s
}

func (p *CodeWriterStruct) AddField(typeName string, fieldName string) {
	if typeName == "" || fieldName == "" {
		plog.Errorf("非法字符串type=%v field=%v\n", typeName, fieldName)
		return
	}
	for _, node := range p.fields {
		if node.fieldName == fieldName {
			plog.Error("field已存在", fieldName)
			return
		}
	}
	p.fields = append(p.fields, CodeNode{
		typeName:  typeName,
		fieldName: fieldName,
	})
}

func (p *CodeWriterStruct) AddMap(keyType string, valueType string, fieldName string) {
	if keyType == "" || valueType == "" || fieldName == "" {
		plog.Errorf("非法字符串key=%v value=%v field=%v\n", keyType, valueType, fieldName)
		return
	}
	for _, node := range p.fields {
		if node.fieldName == fieldName {
			plog.Error("field已存在", fieldName)
			return
		}
	}
	p.fields = append(p.fields, CodeNode{
		fieldEnum: "map",
		keyName:   keyType,
		typeName:  valueType,
		fieldName: fieldName,
	})
}

func (p *CodeWriterStruct) AddSlice(typeName string, fieldName string) {
	if typeName == "" || fieldName == "" {
		plog.Errorf("非法字符串type=%v field=%v\n", typeName, fieldName)
		return
	}
	for _, node := range p.fields {
		if node.fieldName == fieldName {
			plog.Error("field已存在", fieldName)
			return
		}
	}
	p.fields = append(p.fields, CodeNode{
		fieldEnum: "slice",
		typeName:  typeName,
		fieldName: fieldName,
	})
}

// 缩进后面再弄吧
func (p *CodeWriterStruct) Save() {
	p.owner.writeLine(fmt.Sprintf("\n%s[Serializable]", p.prefix))                    // 写入序列化头
	p.owner.writeLine(fmt.Sprintf("\n%spublic class %s {\n", p.prefix, p.structName)) // 写入struct头
	for _, field := range p.fields {
		switch field.fieldEnum {
		case "map":
			p.owner.writeLine(fmt.Sprintf("%s\tpublic Dictionary<%s, %s> %s;\n", p.prefix, field.keyName, field.typeName, field.fieldName)) // 写入map子项

		case "slice":
			p.owner.writeLine(fmt.Sprintf("%s\tpublic List<%s> %s;\n", p.prefix, field.typeName, field.fieldName)) // 写入slice子项

		case "":
			p.owner.writeLine(fmt.Sprintf("%s\tpublic %s %s;\n", p.prefix, field.typeName, field.fieldName)) // 写入普通子项

		default:
			plog.Error("invalid fieldEnum", field.fieldEnum)
			return
		}
	}
	p.owner.writeLine(fmt.Sprintf("%s}\n", p.prefix)) // 写入struct尾
}
