package paladin

import (
	"cmn"
	"frm/plog"
	"reflect"
	"strconv"
	"strings"
)

type StructBuilder struct {
	StructType reflect.Type

	// 内部使用的变量
	typeDesc []string // 类型简单描述. 便于解析
}

// 新建struct创建器
// rows[0] 是数据类型
// rows[1] 是数据名称
// rows[2] 是辅助记忆描述，可忽略
// rows[3] 数据结构描述行，描述是member还是map，还是slice
func NewStructBuilder(rows [][]string) *StructBuilder {
	builder := new(StructBuilder)
	builder.typeDesc = rows[0]
	fields := make([]reflect.StructField, 0, 8)
	for column := 0; column < len(rows[0]); column++ {
		fields = append(fields, reflect.StructField{
			Type: builder.memberType(rows[0][column]),
			Name: cmn.CamelName(rows[1][column]),
		})
	}
	builder.StructType = reflect.StructOf(fields)
	return builder
}

// 根据内容创建实例
// 默认第一列是id
func (p *StructBuilder) CreateInstance(rowIdx int, row []string) (id int, value interface{}) {
	var err error
	if len(row) == 0 {
		return 0, nil
	}

	id, err = strconv.Atoi(row[0])
	if err != nil {
		plog.Error("第一列数据必须要是ID！当前值为", row[0])
		return 0, nil
	}
	structValue := reflect.New(p.StructType)
	elem := structValue.Elem()
	for column := 0; column < len(row); column++ {
		p.assignMember(elem, rowIdx, column, row)
	}
	return 0, structValue.Interface()
}

// 给member成员赋值
func (p *StructBuilder) assignMember(elem reflect.Value, rowIdx int, column int, row []string) {
	defer func() {
		switch err := recover().(type) {
		case nil:

		case error:
			plog.Errorf("读取%d行%d列数据时发生错误%v\n", rowIdx, column, err)

		default:
			plog.Errorf("读取%d行%d列数据时发生错误%v\n", rowIdx, column, err)
		}
	}()

	if row[column] == "NULL" {
		return
	}

	switch p.typeDesc[column] {
	case "INT":
		value, err := strconv.ParseInt(row[column], 10, 64)
		if err != nil {
			plog.Errorf("错误的INT数值%s, 第%d行第%d列\n", row[column], rowIdx, column)
			return
		}
		elem.Field(column).SetInt(value)

	case "UINT":
		value, err := strconv.ParseUint(row[column], 10, 64)
		if err != nil {
			plog.Errorf("错误的UINT数值%s, 第%d行第%d列\n", row[column], rowIdx, column)
			return
		}
		elem.Field(column).SetUint(value)

	case "STRING":
		elem.Field(column).SetString(row[column])

	case "DOUBLE":
		value, err := strconv.ParseFloat(row[column], 64)
		if err != nil {
			plog.Errorf("错误的float64数值%s, 第%d行第%d列\n", row[column], rowIdx, column)
			return
		}
		elem.Field(column).SetFloat(value)

	case "ARRAY":
		plog.Error("to implement!!")

	case "MAP":
		plog.Error("to implement!!")
	}
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
