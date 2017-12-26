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

// 加载枚举表
func (p *XlsxReader) LoadEnum(enums []conf.EnumItem) error {
	return nil
}

// 读取数据
func (p *XlsxReader) Read(tableName string, xlsxFile string) (*XlsxInfo, error) {
	xlsx, err := excelize.OpenFile(xlsxFile)
	if err != nil {
		plog.Errorf("fail to read %s!! %v\n", xlsxFile, err)
		return nil, err
	}

	info := new(XlsxInfo)
	info.TableName = tableName
	plog.Info(tableName, "读取", xlsxFile, ". 子表", xlsx.GetSheetMap())
	for _, name := range xlsx.GetSheetMap() {
		info.Rows[name] = xlsx.GetRows(name)
	}
	return info, nil
}
