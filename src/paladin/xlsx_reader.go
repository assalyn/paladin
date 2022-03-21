package paladin

import (
	"assalyn/paladin/cmn"
	"assalyn/paladin/conf"
	"assalyn/paladin/frm/plog"
	"github.com/360EntSecGroup-Skylar/excelize"
	"strconv"
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

func (p *XlsxReader) Check(xlsx *excelize.File, sheets []string) (err error) {
	defer func() {
		if err := recover(); err != nil {
			plog.Error("panic: ", err)
		}
	}()
	if xlsx == nil {
		plog.Error("xlsx == nil")
		return cmn.ErrBadXlsx
	}
	sheetNames := make([]string, 0, 4)
	for _, sheetName := range xlsx.GetSheetMap() {
		if len(sheets) != 0 {
			for _, paramSheet := range sheets {
				if sheetName == paramSheet {
					sheetNames = append(sheetNames, sheetName)
					break
				}
			}
		} else {
			sheetNames = append(sheetNames, sheetName)
		}
	}
	// 对所有的sheetNames检查结构, 检查name是否完全一致
	if len(sheetNames) == 0 {
		plog.Error("len(sheetNames) == 0")
		return cmn.ErrBadXlsx
	}
	sheetsData := make([][][]string, 0, len(sheetNames))
	for _, sheetName := range sheetNames {
		rows := xlsx.GetRows(sheetName)
		sheetsData = append(sheetsData, rows)
	}
	sentinelSheet := sheetsData[0]
	if len(sentinelSheet) < 4 {
		plog.Error("rows < 4")
		return cmn.ErrBadXlsx
	}
	for sheetIdx := 1; sheetIdx < len(sheetNames); sheetIdx++ {
		for row := 0; row < 4; row++ {
			for col := 0; col < len(sentinelSheet[0]); col++ {
				if sentinelSheet[row][col] != sheetsData[sheetIdx][row][col] {
					plog.Errorf("子表格式不同!! %v[%v][%v] != %v[%v][%v]\n", sheetNames[sheetIdx], row, col, sheetNames[0], row, col)
					err = cmn.ErrBadXlsx
				}
			}
		}
	}
	return err
}

// 读取数据
func (p *XlsxReader) Read(tableName string, xlsxFile string, sheets []string, enums []conf.EnumItem) (*XlsxInfo, error) {
	xlsx, err := excelize.OpenFile(xlsxFile)
	if err != nil {
		return nil, err
	}
	if p.Check(xlsx, sheets) != nil {
		return nil, cmn.ErrBadXlsx
	}

	info := NewXlsxInfo()
	info.TableName = tableName
	info.Enums = enums
	for _, sheet := range xlsx.GetSheetMap() {
		// 如果指定了sheet，则检查sheet. 否则全表查询
		if len(sheets) != 0 {
			found := false
			for i := 0; i < len(sheets); i++ {
				if sheets[i] == sheet {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		var data [][]string
		rows := xlsx.GetRows(sheet)
		// 过滤数据rows，去掉全空
		for i := conf.Cfg.IgnoreLine; i < len(rows); i++ {
			if rows[i][0] == "" {
				rows = rows[:i]
				break
			}
		}
		// 读取数据
		if p.Horizontal {
			if len(rows) == 0 {
				// 没有数据, 直接返回
				continue
			}

			data = make([][]string, len(rows[0]))
			for col := 0; col < len(rows[0]); col++ {
				data[col] = make([]string, len(rows))
			}
			for rowIdx, row := range rows {
				for column, value := range row {
					data[column][rowIdx] = value
				}
			}
		} else {
			data = rows
		}
		info.Rows[sheet] = data

		// 设置nameDict索引. 枚举表和多语言表不需要这种name->id键值对
		if sheet != "enum" && sheet != "locale" {
			nameCol := p.QueryColumn(rows, "name")
			if nameCol == -1 {
				continue // 不需要索引
			}
			idCol := p.QueryColumn(rows, "id")
			if idCol == -1 {
				continue // 不需要索引
			}
			if info.NameDict == nil {
				info.NameDict = make(map[string]string)
			}
			for rowIdx := conf.Cfg.IgnoreLine; rowIdx < len(rows); rowIdx++ {
				_, err := strconv.Atoi(rows[rowIdx][idCol])
				if err != nil {
					plog.Errorf("%s表%s子表[%d][%d]不正确的ID %v! 错误原因：%v\n", tableName, sheet, rowIdx, idCol, rows[rowIdx][idCol], err)
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
