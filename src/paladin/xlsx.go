package paladin

import (
	"conf"
	"frm/plog"

	"github.com/360EntSecGroup-Skylar/excelize"
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
	xlsx, err := excelize.OpenFile(xlsxFile)
	if err != nil {
		plog.Errorf("fail to read %s!! %v\n", xlsxFile, err)
		return nil, err
	}

	info := NewXlsxInfo()
	info.TableName = tableName
	//plog.Info(tableName, "读取", xlsxFile, " 子表", xlsx.GetSheetMap())
	for _, name := range xlsx.GetSheetMap() {
		info.Rows[name] = xlsx.GetRows(name)
	}
	info.Enums = enums
	return info, nil
}
