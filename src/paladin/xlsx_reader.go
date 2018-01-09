package paladin

import (
	"conf"
	"frm/plog"

	"strconv"

	"github.com/tealeg/xlsx"
)

// xlsx相关操作
type XlsxInfo struct {
	TableName string
	Rows      map[string][][]string // 子表-> [行][列]内容
	Enums     []conf.EnumItem       // 枚举表
	NameDict  map[string]string     // name -> id索引表
}

func NewXlsxInfo() *XlsxInfo {
	info := new(XlsxInfo)
	info.Rows = make(map[string][][]string)
	return info
}

type XlsxReader struct {
	AutoId     bool // 自动生成id
	Horizontal bool // 水平读取
}

func NewXlsxReader(autoId bool, horizontal bool) *XlsxReader {
	p := new(XlsxReader)
	p.AutoId = autoId
	p.Horizontal = horizontal
	return p
}

// 读取数据
func (p *XlsxReader) Read(tableName string, xlsxFile string, enums []conf.EnumItem) (*XlsxInfo, error) {
	xlFile, err := xlsx.OpenFile(xlsxFile)
	if err != nil {
		plog.Errorf("fail to read %s!! %v\n", xlsxFile, err)
		return nil, err
	}

	info := NewXlsxInfo()
	info.TableName = tableName
	info.Enums = enums
	for _, sheet := range xlFile.Sheets {
		var rows [][]string
		// 读取数据
		if p.Horizontal {
			if len(sheet.Rows) == 0 {
				// 没有数据, 直接返回
				continue
			}
			rows = make([][]string, len(sheet.Rows[0].Cells))
			for col := 0; col < len(sheet.Rows[0].Cells); col++ {
				rows[col] = make([]string, len(sheet.Rows))
			}
			for rowIdx, row := range sheet.Rows {
				for column, cell := range row.Cells {
					rows[column][rowIdx] = cell.Value
				}
			}
		} else {
			rows = make([][]string, len(sheet.Rows))
			for rowIdx, row := range sheet.Rows {
				rows[rowIdx] = make([]string, len(row.Cells))
				for column, cell := range row.Cells {
					rows[rowIdx][column] = cell.Value
				}
			}
		}
		info.Rows[sheet.Name] = rows

		// 设置nameDict索引. 枚举表和多语言表不需要这种name->id键值对
		if sheet.Name != "enum" && sheet.Name != "locale" {
			nameCol := p.QueryColumn(rows, "name")
			if nameCol == -1 {
				continue // 不需要索引
			}
			idCol := p.QueryColumn(rows, "id")
			if idCol == -1 {
				continue // 不需要索引
			}
			info.NameDict = make(map[string]string)
			for rowIdx := conf.Cfg.IgnoreLine; rowIdx < len(rows); rowIdx++ {
				_, err := strconv.Atoi(rows[rowIdx][idCol])
				if err != nil {
					plog.Errorf("%s表%s子表[%d][%d]不正确的ID %v! 错误原因：%v\n", tableName, sheet.Name, rowIdx, idCol, rows[rowIdx][idCol], err)
					continue
				}
				info.NameDict[rows[rowIdx][nameCol]] = rows[rowIdx][idCol]
			}
		}
	}
	return info, nil
}

// 返回字段name的column
func (p *XlsxReader) QueryColumn(rows [][]string, colName string) int {
	if len(rows) < 2 {
		return -1
	}

	column := 0
	for ; column < len(rows[1]); column++ {
		if rows[1][column] == colName {
			return column
		}
	}
	return -1
}
