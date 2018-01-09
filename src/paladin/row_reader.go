package paladin

import (
	"cmn"
	"fmt"
	"frm/plog"
	"reflect"
	"regexp"
	"strconv"
)

type RowReader struct {
	header [][]string // 头信息
	desc   []string   // 依赖描述信息
	row    []string   // 当前要解析的行
	col    int        // 操作指针
}

// 生成读取器
func NewRowReader(header [][]string, row []string) *RowReader {
	r := new(RowReader)
	r.header = header
	r.desc = header[3]
	r.row = row
	r.col = 0
	return r
}

// 解析数据. 不支持乱序读取
func (p *RowReader) ReadField(fieldName string, t reflect.Type, field reflect.Value) (value reflect.Value, err error) {
	switch field.Kind() {
	case reflect.Struct:
		for i := 0; i < field.NumField(); i++ {
			p.ReadField(t.Field(i).Name, t.Field(i).Type, field.Field(i))
		}
		return field, nil

	case reflect.Map:
		// 获取一簇一簇的数据，然后一个个赋值
		value = reflect.MakeMap(t)
		for {
			key, elemValue, err := p.readMapValue(fieldName, t.Elem())
			if err != nil && err != cmn.ErrEOF {
				plog.Error("读取数据错误", err)
				return value, err
			}
			if err == cmn.ErrEOF {
				break
			}
			value.SetMapIndex(key, elemValue)
		}
		return value, nil

	case reflect.Slice:
		var elemArray []reflect.Value
		for {
			elemValue, err := p.readSliceValue(fieldName, t.Elem())
			if err != nil && err != cmn.ErrEOF {
				plog.Error("读取数据错误", err)
				return value, err
			}
			if err == cmn.ErrEOF {
				break
			}
			elemArray = append(elemArray, elemValue)
		}
		value = reflect.MakeSlice(t, len(elemArray), len(elemArray))
		for i := 0; i < len(elemArray); i++ {
			value.Index(i).Set(elemArray[i])
		}
		return value, nil

	default:
		p.assignMember(field)
	}
	return field, nil
}

func (p *RowReader) readSliceValue(sliceName string, elemType reflect.Type) (reflect.Value, error) {
	value := reflect.New(elemType).Elem()
	if p.col >= len(p.row) {
		return value, cmn.ErrEOF
	}
	allNull := true
	for i := 0; i < value.NumField(); i++ {
		if p.row[p.col+i] != "NULL" {
			allNull = false
		}
	}
	if allNull {
		// 全部成员都为NULL时，代表这个slice没数据
		return value, cmn.ErrEOF
	}
	for i := 0; i < value.NumField(); i++ {
		// 读不出来了
		if p.matchSliceDesc(sliceName) == false {
			return value, cmn.ErrEOF
		}
		p.assignMember(value.Field(i))
	}
	return value, nil
}

func (p *RowReader) readMapValue(mapName string, elemType reflect.Type) (key reflect.Value, value reflect.Value, err error) {
	value = reflect.New(elemType).Elem()
	for i := 0; i < value.NumField(); i++ {
		// 这个数据不是以前那个map结构了
		if p.matchMapDesc(mapName) == false {
			return key, value, cmn.ErrEOF
		}
		// 键值为NULL时过滤掉这条map数据
		if i == 0 && p.row[p.col] == "NULL" {
			p.col += value.NumField()
			return key, value, cmn.ErrEOF
		}
		p.assignMember(value.Field(i))
	}
	return value.Field(0), value, nil
}

// 给member成员赋值
func (p *RowReader) assignMember(elem reflect.Value) {
	defer func() {
		switch err := recover().(type) {
		case nil:

		case error:
			plog.Errorf("读取%d列数据时发生错误%v\n", p.col, err)

		default:
			plog.Errorf("读取%d列数据时发生错误%v\n", p.col, err)
		}
	}()

	if p.row[p.col] == "NULL" {
		p.col++
		return
	}

	switch elem.Type().Kind() {
	case reflect.Int:
		value, err := strconv.ParseInt(p.row[p.col], 10, 64)
		if err != nil {
			plog.Errorf("错误的INT数值%s, 第%d列\n", p.row[p.col], p.col)
			return
		}
		elem.SetInt(value)

	case reflect.Uint:
		value, err := strconv.ParseUint(p.row[p.col], 10, 64)
		if err != nil {
			plog.Errorf("错误的UINT数值%s, 第%d列\n", p.row[p.col], p.col)
			return
		}
		elem.SetUint(value)

	case reflect.String:
		elem.SetString(p.row[p.col])

	case reflect.Float64:
		value, err := strconv.ParseFloat(p.row[p.col], 64)
		if err != nil {
			plog.Errorf("错误的float64数值%s, 第%d列\n", p.row[p.col], p.col)
			return
		}
		elem.SetFloat(value)
	}
	p.col++
}

// 是否匹配 [rate]#xxx 或
func (p *RowReader) matchSliceDesc(dictName string) bool {
	if p.col >= len(p.row) {
		return false
	}
	matched, _ := regexp.Match(fmt.Sprintf("(?i:^\\[%s\\])", dictName), []byte(p.desc[p.col]))
	return matched
}

// 是否匹配 {rate}#xxx 或
func (p *RowReader) matchMapDesc(dictName string) bool {
	if p.col >= len(p.row) {
		return false
	}
	matched, _ := regexp.Match(fmt.Sprintf("(?i:^\\{%s\\})", dictName), []byte(p.desc[p.col]))
	return matched
}
