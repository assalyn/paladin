package paladin

import (
	"cmn"
	"fmt"
	"frm/plog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type StructBuilder struct {
	// 生成结构
	StructType reflect.Type

	// 内部使用的变量
	typeDesc  []string // 类型简单描述. 便于解析
	layerDesc []string // 层级描述 rows[3]
	rows      [][]string

	// 临时数据
	prevLayerDesc string
}

// 新建struct创建器
// rows[0] 是数据类型
// rows[1] 是数据名称
// rows[2] 是辅助记忆描述，可忽略
// rows[3] 数据结构描述行，描述是member还是map，还是slice
func NewStructBuilder(rows [][]string) *StructBuilder {
	builder := new(StructBuilder)
	builder.typeDesc = rows[0]
	builder.layerDesc = rows[3]
	builder.rows = rows
	return builder
}

func (p *StructBuilder) BuildStruct() {
	fields := make([]reflect.StructField, 0, 8)
	column := 0
	for column < len(p.layerDesc) {
		// 获取数据簇
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

// 根据一行数据创建实例, 默认第一列是id
func (p *StructBuilder) CreateInstance(row []string) (id int, value interface{}, err error) {
	if len(row) == 0 {
		return 0, nil, cmn.ErrFail
	}

	// 解析id
	id, err = strconv.Atoi(row[0])
	if err != nil {
		plog.Error("第一列数据必须要是ID！当前值为", row[0])
		return 0, nil, cmn.ErrFail
	}

	// field赋值
	structValue := reflect.New(p.StructType)
	elem := structValue.Elem() // 这是对象
	reader := NewRowReader(p.rows[:4], row)
	for i := 0; i < elem.NumField(); i++ {
		value, err := reader.ReadField(p.StructType.Field(i).Name, p.StructType.Field(i).Type, elem.Field(i))
		if err != nil {
			plog.Error("fail to assign field!!", err)
			return 0, nil, cmn.ErrFail
		}
		elem.Field(i).Set(value)
	}
	return id, structValue.Interface(), nil
}

// [XXX]数组结构; {XXX}map结构; XXX内部数据结构
func (p *StructBuilder) parseField(column *int) (field reflect.StructField, err error) {
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

// 获取数据簇
func (p *StructBuilder) getFieldCluster(startCol int) (endCol int) {
	// 同一个结构要留下来
	i := startCol
	for ; i < len(p.layerDesc); i++ {
		if strings.Contains(p.layerDesc[i], p.layerDesc[startCol]) == false {
			return i
		}
	}
	return i
}

// 成员类型
func (p *StructBuilder) memberType(typeName string) reflect.Type {
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

// 通过layer描述获取数据结构
func (p *StructBuilder) layerType(layerDesc string) (subName string, subType string) {
	// 过滤掉#后面的部分
	rex, _ := regexp.Compile("#.*$")
	cleanStr := rex.ReplaceAllString(layerDesc, "")

	if cleanStr == "" {
		return "", "member"
	} else if cleanStr[0] == '[' && cleanStr[len(cleanStr)-1] == ']' {
		return cleanStr[1 : len(cleanStr)-1], "slice"
	} else if cleanStr[0] == '{' && cleanStr[len(cleanStr)-1] == '}' {
		return cleanStr[1 : len(cleanStr)-1], "map"
	}
	matched, _ := regexp.Match("[a-zA-Z0-9_.]*", []byte(cleanStr))
	if matched {
		return cleanStr, "struct"
	} else {
		return "", "invalid"
	}
}
