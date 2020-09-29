package paladin

// 中间数据初始化代码

import "assalyn/paladin/frm/plog"

// 构建枚举替换结构
func (p *Parser) structEnumSwapMap() {
	// 解析enum表
	for subTable, rows := range p.Enum.Rows {
		// 枚举表的前三行忽略
		swapTable := make(map[string]string)
		for rowIdx, row := range rows {
			if rowIdx < 3 {
				continue
			}
			if len(row) < 2 {
				plog.Errorf("枚举表异常，数据长度=%v 第%v行\n", len(row), rowIdx)
				continue
			}
			swapTable[row[0]] = row[1]
		}
		p.EnumSwapDict[subTable] = swapTable
	}
}

// 构建参数展开结构
func (p *Parser) structParamUnfoldMap() {

}

// 构建多语言替换结构 : locale -> field -> content
func (p *Parser) structLocaleSwapMap() {
	for subTable, rows := range p.Locale.Rows {
		localeSwapTable := make(map[string]map[string]string)
		// rows[1] 是语言描述栏. cn, en
		if len(rows[1]) < 2 {
			plog.Error("错误多语言数据，数据表列<2", subTable)
			continue
		}
		// 语言从第1列开始，第0列是alias
		for column := 1; column < len(rows[1]); column++ {
			swapTable := make(map[string]string)
			for rowIdx, row := range rows {
				if rowIdx < 3 {
					continue
				}
				swapTable[row[0]] = row[column]
			}
			localeSwapTable[rows[1][column]] = swapTable
		}
		p.LocaleSwapDict[subTable] = localeSwapTable
	}
}
