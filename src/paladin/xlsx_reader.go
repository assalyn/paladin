package paladin

import (
	"conf"
	"frm/plog"

	"github.com/tealeg/xlsx"
)

// xlsx相关操作
type XlsxInfo struct {
	TableName string
	Rows      map[string][][]string // 子表-> [行][列]内容
	Enums     []conf.EnumItem       // 枚举表
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
		if p.Horizontal {
			if len(sheet.Rows) == 0 {
				// 没有数据, 直接返回
				continue
			}
			l := make([][]string, len(sheet.Rows[0].Cells))
			for col := 0; col < len(sheet.Rows[0].Cells); col++ {
				l[col] = make([]string, len(sheet.Rows))
			}
			for rowIdx, row := range sheet.Rows {
				for column, cell := range row.Cells {
					l[column][rowIdx] = cell.Value
				}
			}
			info.Rows[sheet.Name] = l
		} else {
			l := make([][]string, len(sheet.Rows))
			for rowIdx, row := range sheet.Rows {
				l[rowIdx] = make([]string, len(row.Cells))
				for column, cell := range row.Cells {
					l[rowIdx][column] = cell.Value
				}
			}
			info.Rows[sheet.Name] = l
		}
	}
	return info, nil
}
