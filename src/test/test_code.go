package main

import (
	"assalyn/paladin/frm/plog"
	"assalyn/paladin/paladin"
	"fmt"
	"github.com/tealeg/xlsx"
	"reflect"
	"regexp"
	"strings"
)

func main() {
	plog.InitLog("test_code.log", plog.LOG_TRACE)

	// 读取数据
	data := readStruct()
	builder := NewTestBuilder(data[:4])
	builder.BuildStruct()
	fmt.Printf("%#v\n", builder.StructType)
	builder.CreateInstance()

	coder := paladin.NewGoCodeBuilder("AAA", "", "")
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

	prevLayerDesc string
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
		// 空数据结构，略过，可能是重复结构
		if field.Name == "" {
			continue
		}
		fields = append(fields, field)
	}
	p.StructType = reflect.StructOf(fields)
}

// [XXX]数组结构; {XXX}map结构; XXX内部数据结构
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

	// 同一个结构要留下来
	sentry := *column
	for ; sentry < len(p.layerDesc); sentry++ {
		if strings.Contains(p.layerDesc[sentry], currDesc) == false {
			break
		}
	}

	// 将连续的一组都选出来. 以currDesc为基准进行解析
	subName, subType := p.layerType(currDesc)
	// 如果一样，说明是同一实例
	if subName == p.prevLayerDesc {
		*column = sentry
		return
	}
	// 如果是prevLayerDesc.yyy, 说明存在继承关系
	switch subType {
	case "member":
		// 子成员
		field = reflect.StructField{
			Type: p.memberType(p.rows[0][*column]),
			Name: cmn.CamelName(p.rows[1][*column]),
		}
		*column++

	case "struct":
		// 子struct类型, struct结束时创建数据结构
		var fs []reflect.StructField
		for j := *column; j < sentry; j++ {
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
		*column = sentry

	case "slice":
		// 记录这个slice类型, 结束时创建数据结构. 并赋值给上级数据结构. 要用递归结构
		var fs []reflect.StructField
		for j := *column; j < sentry; j++ {
			fs = append(fs, reflect.StructField{
				Type: p.memberType(p.rows[0][j]),
				Name: cmn.CamelName(p.rows[1][j]),
			})
		}
		subStruct := reflect.StructOf(fs)
		sliceStruct := reflect.SliceOf(subStruct)
		field = reflect.StructField{
			Type: sliceStruct,
			Name: cmn.CamelName(subName),
		}
		*column = sentry

	case "map":
		// 记录这个map类型, 结束时创建数据结构
		var fs []reflect.StructField
		for j := *column; j < sentry; j++ {
			fs = append(fs, reflect.StructField{
				Type: p.memberType(p.rows[0][j]),
				Name: cmn.CamelName(p.rows[1][j]),
			})
		}
		subStruct := reflect.StructOf(fs)
		sliceStruct := reflect.MapOf(fs[0].Type, subStruct)
		field = reflect.StructField{
			Type: sliceStruct,
			Name: cmn.CamelName(subName),
		}
		*column = sentry

	default:
		fmt.Println("错误的依赖类型", currDesc)
		return
	}

	p.prevLayerDesc = subName
	return
}

func (p *TestBuilder) memberType(typeName string) reflect.Type {
	typeName = strings.ToUpper(typeName)
	switch typeName {
	case "BOOL":
		return reflect.TypeOf(true)

	case "UINT":
		return reflect.TypeOf(uint(0))

	case "INT":
		return reflect.TypeOf(int(0))

	case "INT32":
		return reflect.TypeOf(int32(0))

	case "STRING":
		return reflect.TypeOf("")

	case "FLOAT":
		return reflect.TypeOf(float32(0))

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
	// 过滤掉#后面的部分
	rex, _ := regexp.Compile("#.*$")
	cleanStr := rex.ReplaceAllString(layerDesc, "")

	if cleanStr == "" {
		return "", "member"
	} else if cleanStr[0] == '[' && cleanStr[len(cleanStr)-1] == ']' {
		return cleanStr[1 : len(cleanStr)-1], "slice"
	} else if cleanStr[0] == '{' && cleanStr[len(cleanStr)-1] == '}' {
		return cleanStr[1 : len(cleanStr)-1], "map"
	} else if cleanStr[0] == '<' && cleanStr[len(cleanStr)-1] == '>' {
		return cleanStr[1 : len(cleanStr)-1], "struct"
	}
	return "", "invalid"
}
