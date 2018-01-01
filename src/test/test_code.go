package main

import (
	"frm/plog"

	"cmn"
	"reflect"

	"strings"

	"fmt"

	"regexp"

	"paladin"

	"github.com/tealeg/xlsx"
)

func main() {
	// 读取数据
	data := readStruct()
	builder := NewTestBuilder(data[:4])
	builder.BuildStruct()
	fmt.Printf("%#v\n", builder.StructType)
	builder.CreateInstance()

	coder := paladin.NewCodeBuilder("AAA")
	coder.DebugType(builder.StructType, "AAA")
}

func readStruct() [][]string {
	xlFile, err := xlsx.OpenFile("bin/xlsx/位置表.xlsx")
	if err != nil {
		plog.Errorf("fail to read %s!! %v\n", "bin/xlsx/位置表.xlsx", err)
		return nil
	}

	for _, sheet := range xlFile.Sheets {
		l := make([][]string, len(sheet.Rows))
		for rowIdx, row := range sheet.Rows {
			l[rowIdx] = make([]string, len(row.Cells))
			for column, cell := range row.Cells {
				l[rowIdx][column] = cell.Value
			}
		}
		return l
	}
	return nil
}

// 测试生成程序
type TestBuilder struct {
	typeDesc   []string // 类型描述 rows[0]
	layerDesc  []string // 层级描述 rows[3]
	rows       [][]string
	StructType reflect.Type
}

func NewTestBuilder(rows [][]string) *TestBuilder {
	builder := new(TestBuilder)
	builder.typeDesc = rows[0]
	builder.layerDesc = rows[3]
	builder.rows = rows
	fmt.Println("layer description=", builder.layerDesc)
	return builder
}

func (p *TestBuilder) BuildStruct() {
	fields := make([]reflect.StructField, 0, 8)
	column := 0
	for column < len(p.layerDesc) {
		field, err := p.parseField(&column)
		if err != nil {
			fmt.Println("错误的数据结构", err)
			return
		}
		// 空数据结构，退出
		if field.Name == "" {
			break
		}
		fields = append(fields, field)
	}
	p.StructType = reflect.StructOf(fields)
}

// 先不考虑map
func (p *TestBuilder) parseField(column *int) (field reflect.StructField, err error) {
	currDesc := p.layerDesc[*column]
	if currDesc == "" {
		field = reflect.StructField{
			Type: p.memberType(p.rows[0][*column]),
			Name: cmn.CamelName(p.rows[1][*column]),
		}
		*column++
		return
	}

	i := *column
	for ; i < len(p.layerDesc); i++ {
		if p.layerDesc[i] != currDesc {
			break
		}
	}

	// 将连续的一组都选出来
	layers := strings.Split(currDesc, ".")
	// [XXX]数组结构; {XXX}map结构; XXX内部数据结构
	subName, subType := p.layerType(layers[0])
	switch subType {
	case "member":
		// 子成员
		field = reflect.StructField{
			Type: p.memberType(p.rows[0][*column]),
			Name: cmn.CamelName(p.rows[1][*column]),
		}
		*column++
		return

	case "struct":
		// 子struct类型, struct结束时创建数据结构
		var fs []reflect.StructField
		for j := *column; j < i; j++ {
			fs = append(fs, reflect.StructField{
				Type: p.memberType(p.rows[0][j]),
				Name: cmn.CamelName(p.rows[1][j]),
			})
		}
		subStruct := reflect.StructOf(fs)
		field = reflect.StructField{
			Type: subStruct,
			Name: cmn.CamelName(subName),
		}
		*column = i
		return

	case "slice":
		// 记录这个slice类型, 结束时创建数据结构. 并赋值给上级数据结构. 要用递归结构
		for j := *column; j < i; j++ {

		}
		subStruct := reflect.StructOf()
		sliceStruct := reflect.SliceOf(subStruct)
		field = reflect.StructField{
			Type: sliceStruct,
			Name: cmn.CamelName(subName),
		}
		*column = i
		return

	case "map":
		// 记录这个map类型, 结束时创建数据结构

	default:
		fmt.Println("错误的依赖类型", currDesc)
	}
	return
}

func (p *TestBuilder) memberType(typeName string) reflect.Type {
	typeName = strings.ToUpper(typeName)
	switch typeName {
	case "UINT":
		return reflect.TypeOf(uint(0))

	case "INT":
		return reflect.TypeOf(int(0))

	case "STRING":
		return reflect.TypeOf("")

	case "DOUBLE":
		return reflect.TypeOf(float64(0))

	default:
		plog.Error("Unsupport type!", typeName)
		return reflect.TypeOf(int(0))
	}
}

func (p *TestBuilder) CreateInstance() {

}

// 通过layer描述获取数据结构
func (p *TestBuilder) layerType(layerDesc string) (subName string, subType string) {
	if layerDesc == "" {
		return "", "member"
	} else if layerDesc[0] == '[' && layerDesc[len(layerDesc)-1] == ']' {
		return layerDesc[1 : len(layerDesc)-2], "slice"
	} else if layerDesc[0] == '{' && layerDesc[len(layerDesc)-1] == '}' {
		return layerDesc[1 : len(layerDesc)-2], "map"
	}
	matched, _ := regexp.Match("[a-zA-Z0-9_]*", []byte(layerDesc))
	if matched {
		return layerDesc, "struct"
	} else {
		return "", "invalid"
	}
}
