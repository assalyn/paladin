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

func (p *StructBuilder) BuildStruct() error {
	fields := make([]reflect.StructField, 0, 8)
	subName := ""
	column := 0
	for column < len(p.layerDesc) {
		// 获取数据簇
		field, err := p.parseField(0, &column, &subName)
		if err != nil {
			if err == cmn.ErrSkip {
				continue
			} else {
				plog.Error("错误的数据结构", err)
			}
			return err
		}
		fields = append(fields, field)
	}
	p.StructType = reflect.StructOf(fields)
	return nil
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
		if err != nil && err != cmn.ErrNull && err != cmn.ErrSkip {
			plog.Error("fail to assign field!!", err)
			return 0, nil, cmn.ErrFail
		}
		elem.Field(i).Set(value)
	}
	return id, structValue.Interface(), nil
}

// [XXX]数组结构; {XXX}map结构; XXX内部数据结构
func (p *StructBuilder) parseField(descSkip int, column *int, prevSubName *string) (field reflect.StructField, err error) {
	//fmt.Printf("parseField descSkip=%v column=%v\n", descSkip, *column)
	currDesc := p.layerDesc[*column][descSkip:]
	if currDesc == "" {
		field = reflect.StructField{
			Type: p.memberType(0, *column),
			Name: cmn.CamelName(p.rows[1][*column]),
		}
		*column++
		return field, nil
	}

	// 同一个结构要留下来
	sentry := *column
	for ; sentry < len(p.layerDesc); sentry++ {
		if strings.Index(p.layerDesc[sentry], p.layerDesc[*column]) != 0 {
			break
		}
	}

	// 将连续的一组都选出来. 以currDesc为基准进行解析
	subName, subType := p.layerType(currDesc)
	if subName == *prevSubName {
		// 如果一样，说明是同一结构的不同表达，比如[cast]#1,[cast]#2. 这种情况下，[cast]#2不用需要再处理一次了, 因为结构和[cast]#1是一样的
		*column = sentry
		return field, cmn.ErrSkip
	}
	switch subType {
	case "member":
		// 子成员
		field = reflect.StructField{
			Type: p.memberType(0, *column),
			Name: cmn.CamelName(p.rows[1][*column]),
		}
		*column++

	case "struct":
		field, err = p.parseFieldStruct(subName, column, sentry)

	case "slice":
		field, err = p.parseFieldSlice(subName, column, sentry)

	case "map":
		field, err = p.parseFieldMap(subName, column, sentry)

	default:
		fmt.Println("错误的依赖类型", currDesc)
		return field, cmn.ErrFail
	}

	*prevSubName = subName
	return field, err
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
func (p *StructBuilder) memberType(rowIdx int, column int) reflect.Type {
	typeName := strings.ToUpper(p.rows[rowIdx][column])
	switch typeName {
	case "UINT":
		return reflect.TypeOf(uint(0))

	case "INT":
		return reflect.TypeOf(int(0))

	case "INT32":
		return reflect.TypeOf(int32(0))

	case "INT64":
		return reflect.TypeOf(int64(0))

	case "STRING":
		return reflect.TypeOf("")

	case "DOUBLE":
		return reflect.TypeOf(float64(0))

	default:
		plog.Errorf("Unsupport type!rows[%v][%v] = %v\n", rowIdx, column, typeName)
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
	} else if cleanStr[0] == '<' && cleanStr[len(cleanStr)-1] == '>' {
		return cleanStr[1 : len(cleanStr)-1], "struct"
	}
	return "", "invalid"
}

func (p *StructBuilder) parseFieldSlice(subName string, column *int, sentry int) (field reflect.StructField, err error) {
	var fs []reflect.StructField
	var sfield reflect.StructField
	var parentHdrLen = p.SliceTailIdx(p.layerDesc[*column], subName)
	var prevSubName = ""

	for j := *column; j < sentry; {
		sfield, err = p.parseField(parentHdrLen, &j, &prevSubName)
		if err != nil {
			if err == cmn.ErrSkip {
				continue
			} else {
				return field, err
			}
		}
		fs = append(fs, sfield)
	}
	var subStruct reflect.Type
	if len(fs) > 1 {
		subStruct = reflect.StructOf(fs)
	} else {
		subStruct = fs[0].Type
	}
	sliceStruct := reflect.SliceOf(subStruct)
	field = reflect.StructField{
		Type: sliceStruct,
		Name: cmn.CamelName(subName),
	}
	if strings.EqualFold(field.Name, subName) == false {
		plog.Error("slice只能使用Camel命名, ", subName, "!=", field.Name)
		return field, cmn.ErrEOF
	}
	*column = sentry
	return field, nil
}

func (p *StructBuilder) parseFieldMap(subName string, column *int, sentry int) (field reflect.StructField, err error) {
	var fs []reflect.StructField
	var sfield reflect.StructField
	var parentHdrLen = p.MapTailIdx(p.layerDesc[*column], subName)
	var prevSubName = ""

	for j := *column; j < sentry; {
		sfield, err = p.parseField(parentHdrLen, &j, &prevSubName)
		if err != nil {
			return field, err
		}
		fs = append(fs, sfield)
	}
	var subStruct reflect.Type
	if len(fs) > 1 {
		subStruct = reflect.StructOf(fs)
	} else {
		subStruct = fs[0].Type
	}
	mapStruct := reflect.MapOf(fs[0].Type, subStruct)
	field = reflect.StructField{
		Type: mapStruct,
		Name: cmn.CamelName(subName),
	}
	if strings.EqualFold(field.Name, subName) == false {
		plog.Error("map只能使用Camel命名, ", subName, "!=", field.Name)
		return field, cmn.ErrEOF
	}
	*column = sentry
	return field, nil
}

func (p *StructBuilder) parseFieldStruct(subName string, column *int, sentry int) (field reflect.StructField, err error) {
	var fs []reflect.StructField
	var sfield reflect.StructField
	var parentHdrLen = p.StructTailIdx(p.layerDesc[*column], subName)
	var prevSubName = ""

	for j := *column; j < sentry; {
		sfield, err = p.parseField(parentHdrLen, &j, &prevSubName)
		if err != nil {
			return field, err
		}
		fs = append(fs, sfield)
	}
	subStruct := reflect.StructOf(fs)
	field = reflect.StructField{
		Type: subStruct,
		Name: cmn.CamelName(subName),
	}
	*column = sentry
	return field, nil
}

func (p *StructBuilder) SliceTailIdx(name string, subTypeName string) int {
	fullSubName := fmt.Sprintf("[%v]", subTypeName)
	idx := strings.Index(name, fullSubName)
	return idx + len(fullSubName) + 2
}

func (p *StructBuilder) MapTailIdx(name string, subTypeName string) int {
	fullSubName := fmt.Sprintf("{%v}", subTypeName)
	idx := strings.Index(name, fullSubName)
	return idx + len(fullSubName) + 2
}

// 找到子类型尾巴
func (p *StructBuilder) StructTailIdx(name string, subTypeName string) int {
	fullSubName := fmt.Sprintf("<%v>", subTypeName)
	idx := strings.Index(name, fullSubName)
	return idx + len(fullSubName)
}
