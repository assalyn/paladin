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
	if p.col >= len(p.row) {
		return value, cmn.ErrEOF
	}
	switch field.Kind() {
	case reflect.Struct:
		for i := 0; i < field.NumField(); i++ {
			_, err = p.ReadField(t.Field(i).Name, t.Field(i).Type, field.Field(i))
			if err == nil {
				// 什么都不做
			} else if err == cmn.ErrSkip || err == cmn.ErrNull {
				continue
			} else if err == cmn.ErrEOF {
				break
			} else {
				plog.Error("读取数据错误", err)
				return value, err
			}
		}
		return field, nil

	case reflect.Map:
		// 获取一簇一簇的数据，然后一个个赋值
		value = reflect.MakeMap(t)
		for {
			originCol := p.col
			key, elemValue, err := p.readMapValue(fieldName, t.Elem())
			if err == nil {
				// allNull时不添加
				allNull := true
				for i := originCol; i < p.col; i++ {
					if p.row[i] != "NULL" {
						allNull = false
						break
					}
				}
				if !allNull {
					value.SetMapIndex(key, elemValue)
				}
			} else if err == cmn.ErrEOF {
				break
			} else if err == cmn.ErrNull {
				continue
			} else {
				plog.Error("读取数据错误", err)
				return value, err
			}

		}
		return value, nil

	case reflect.Slice:
		var elemArray []reflect.Value
		for {
			originCol := p.col
			elemValue, err := p.readSliceValue(fieldName, t.Elem())
			if err == nil {
				// allNull时不添加
				allNull := true
				for i := originCol; i < p.col; i++ {
					if p.row[i] != "NULL" {
						allNull = false
						break
					}
				}
				if !allNull {
					elemArray = append(elemArray, elemValue)
				}
			} else if err == cmn.ErrEOF {
				break
			} else if err == cmn.ErrNull {
				continue
			} else {
				plog.Error("读取数据错误", err)
				return value, err
			}
		}
		value = reflect.MakeSlice(t, len(elemArray), len(elemArray))
		for i := 0; i < len(elemArray); i++ {
			value.Index(i).Set(elemArray[i])
		}
		//fmt.Printf("ReadField return: %v %#v\n", fieldName, value)
		return value, nil

	default:
		if err = p.assignMember(field); err != nil {
			return field, err
		}
		return field, nil
	}
}

// 内部还可能是复杂结构啊...比如slice, map
func (p *RowReader) readSliceValue(sliceName string, elemType reflect.Type) (reflect.Value, error) {
	value := reflect.New(elemType).Elem()
	if p.col >= len(p.row) {
		return value, cmn.ErrEOF
	}
	if elemType.Kind() == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			if p.matchSliceDesc(sliceName) == false {
				return value, cmn.ErrEOF
			}
			v, err := p.ReadField(elemType.Field(i).Name, elemType.Field(i).Type, value.Field(i))
			if err == nil {
				value.Field(i).Set(v)
			} else if err == cmn.ErrSkip || err == cmn.ErrNull {
				continue
			} else if err == cmn.ErrEOF {
				break
			} else {
				plog.Error("读取数据错误", err)
				return value, err
			}
		}
	} else {
		if p.matchSliceDesc(sliceName) == false {
			return value, cmn.ErrEOF
		}
		if err := p.assignMember(value); err != nil {
			return value, err
		}
	}
	return value, nil
}

func (p *RowReader) readMapValue(mapName string, elemType reflect.Type) (key reflect.Value, value reflect.Value, err error) {
	value = reflect.New(elemType).Elem()
	if p.col >= len(p.row) {
		return key, value, cmn.ErrEOF
	}
	if elemType.Kind() == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			if p.matchMapDesc(mapName) == false {
				return key, value, cmn.ErrEOF
			}
			v, err := p.ReadField(elemType.Field(i).Name, elemType.Field(i).Type, value.Field(i))
			if err == nil {
				value.Field(i).Set(v)
			} else if err == cmn.ErrSkip || err == cmn.ErrNull {
				continue
			} else if err == cmn.ErrEOF {
				break
			} else {
				plog.Error("读取数据错误", err)
				return key, value, err
			}
		}
	} else {
		// 这个数据不是以前那个map结构了
		if p.matchMapDesc(mapName) == false {
			return key, value, cmn.ErrEOF
		}
		if err := p.assignMember(value); err != nil {
			return key, value, err
		}
	}
	return value.Field(0), value, nil
}

// 给member成员赋值
func (p *RowReader) assignMember(elem reflect.Value) error {
	col := p.col
	p.col++
	if p.row[col] == "NULL" {
		return cmn.ErrNull
	}

	switch elem.Type().Kind() {
	case reflect.Int:
		value, err := strconv.ParseInt(p.row[col], 10, 64)
		if err != nil {
			plog.Errorf("错误的INT数值%s, 第%d列\n", p.row[col], col)
			return cmn.ErrFail
		}
		elem.SetInt(value)

	case reflect.Uint:
		value, err := strconv.ParseUint(p.row[col], 10, 64)
		if err != nil {
			plog.Errorf("错误的UINT数值%s, 第%d列\n", p.row[col], col)
			return cmn.ErrFail
		}
		elem.SetUint(value)

	case reflect.String:
		elem.SetString(p.row[col])

	case reflect.Float64:
		value, err := strconv.ParseFloat(p.row[col], 64)
		if err != nil {
			plog.Errorf("错误的float64数值%s, 第%d列\n", p.row[col], col)
			return cmn.ErrFail
		}
		elem.SetFloat(value)
	}
	return nil
}

// 是否匹配 [rate]#xxx 或
// todo 有缺陷，在解析子slice时，没办法区分上级。比如[cast]#1[effect]#1和[cast]#2[effect]#1
func (p *RowReader) matchSliceDesc(sliceName string) bool {
	if p.col >= len(p.row) {
		return false
	}
	matched, _ := regexp.Match(fmt.Sprintf("(?i:\\[%s\\])", sliceName), []byte(p.desc[p.col]))
	return matched
}

// 是否匹配 {rate}#xxx 或
func (p *RowReader) matchMapDesc(dictName string) bool {
	if p.col >= len(p.row) {
		return false
	}
	matched, _ := regexp.Match(fmt.Sprintf("(?i:\\{%s\\})", dictName), []byte(p.desc[p.col]))
	return matched
}
