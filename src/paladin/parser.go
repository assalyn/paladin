package paladin

import (
	"assalyn/paladin/cmn"
	"assalyn/paladin/conf"
	"assalyn/paladin/frm/plog"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	// 数据
	Enum   *XlsxInfo
	Locale *XlsxInfo            // 多语言
	Xlsx   map[string]*XlsxInfo // 表代号->xlsx数据

	// 中间数据
	EnumSwapDict    map[string]map[string]string            // 枚举替换表 子表->field->内容
	ParamUnfoldDict map[string]map[string]string            // 参数展开表 子表->替换项->内容
	LocaleSwapDict  map[string]map[string]map[string]string // 多语言替换表 子表->语言->field->内容
	Output          map[string]map[int]interface{}          // 待输出的数据 子表->id->内容

	// 内部使用变量
	outputDir string // 输出目录
	stubDir   string // 桩代码目录
	localeDir string // 多语言目录
	genGolang bool   // 是否生成golang桩代码
	genCsharp bool   // 是否生成csharp桩代码
}

func NewParser(outputDir string, stubDir string, localeDir string, genGolang bool, genCsharp bool) *Parser {
	p := new(Parser)
	p.Xlsx = make(map[string]*XlsxInfo)

	p.EnumSwapDict = make(map[string]map[string]string)
	p.ParamUnfoldDict = make(map[string]map[string]string)
	p.LocaleSwapDict = make(map[string]map[string]map[string]string)
	p.Output = make(map[string]map[int]interface{})

	p.outputDir = outputDir
	p.stubDir = stubDir
	p.localeDir = localeDir
	p.genGolang = genGolang
	p.genCsharp = genCsharp
	return p
}

func (p *Parser) Start() {
	// 加载文件
	p.loadFiles()

	// 准备中间数据
	p.prepare()

	// 解析数据文件
	p.parse()

	// 导出为json
	p.output("json")

	// 输出多语言文件
	p.outputLocale()

	// 是否生成桩文件
	p.genStubCode()
}

// 加载文件
func (p *Parser) loadFiles() {
	// 加载枚举文件
	reader := NewXlsxReader(false, false)
	enumInfo, err := reader.Read("enum", conf.Cfg.EnumFile, nil, nil)
	if err != nil {
		plog.Panic("读取枚举表失败!!", err)
	}
	p.Enum = enumInfo

	// 加载xlsx文件
	for tblName, tbl := range conf.Cfg.Tables {
		reader := NewXlsxReader(tbl.AutoId, tbl.Horizontal)
		info, err := reader.Read(tblName, tbl.Workbook, tbl.Sheet, tbl.Enums)
		if err != nil {
			plog.Errorf("读取%s.%s失败！错误码:%v\n", tbl.Workbook, tbl.Sheet, err)
		}
		p.Xlsx[tblName] = info
	}

	// 加载多语言文件
	reader = NewXlsxReader(false, false)
	localeInfo, err := reader.Read("locale", conf.Cfg.LocaleFile, nil, nil)
	if err != nil {
		plog.Info("无多语言表，不进行多语言处理")
	}
	p.Locale = localeInfo
}

func (p *Parser) prepare() {
	// 构建枚举替换结构
	p.structEnumSwapMap()

	// 构建参数展开结构
	p.structParamUnfoldMap()

	// 构建多语言替换结构
	p.structLocaleSwapMap()
}

// 解析
func (p *Parser) parse() {
	for tableName, info := range p.Xlsx {
		p.parseXlsx(tableName, info)
	}
}

// 导出
func (p *Parser) output(fmt string) {
	// 校验outputDir是否存在
	if err := os.MkdirAll(p.outputDir, 0777); err != nil {
		plog.Errorf("创建目录%v失败%v\n", p.outputDir, err)
		return
	}
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "json":
		p.outputJson()

	default:
		plog.Panic("Invalid output fmt", fmt)
	}
}

func (p *Parser) outputLocale() {
	// 校验localeDir是否存在
	if err := os.MkdirAll(p.localeDir, 0777); err != nil {
		plog.Errorf("创建目录%v失败%v\n", p.localeDir, err)
		return
	}
	// 导出locale文件
	for tableName, localeSwapDict := range p.LocaleSwapDict {
		for locale, swapTable := range localeSwapDict {
			if err := os.MkdirAll(p.localeDir+"/"+locale, 0777); err != nil {
				plog.Errorf("创建多语言目录%v失败\n", p.localeDir)
				continue
			}
			localeFile, err := os.Create(p.localeDir + "/" + locale + "/" + tableName + ".json")
			if err != nil {
				plog.Error(tableName, "生成多语言文件失败", err)
				continue
			}
			encoder := json.NewEncoder(localeFile)
			if err = encoder.Encode(swapTable); err != nil {
				plog.Error(tableName, "导出多语言文件json失败", err)
				continue
			}
		}
	}
}

// 生成桩文件
func (p *Parser) genStubCode() {
	if p.genGolang {
		var goDir = p.stubDir + "/go/"
		if err := os.MkdirAll(goDir, 0777); err != nil {
			plog.Errorf("创建目录%v失败%v\n", p.stubDir, err)
			return
		}
		p.genGolangStub(goDir)
	}
	if p.genCsharp {
		var csDir = p.stubDir + "/cs/"
		if err := os.MkdirAll(csDir, 0777); err != nil {
			plog.Errorf("创建目录%v失败%v\n", p.stubDir, err)
			return
		}
		p.genCsharpStub(csDir)
	}
}

////////////////////////////////////// 子函数 //////////////////////////////////////
// 解析xlsx文件
func (p *Parser) parseXlsx(tableName string, info *XlsxInfo) {
	xlsxConf := conf.Cfg.Tables[tableName]
	totalStructs := make(map[int]interface{})
	// 创建数据结构
	for subTableName, rows := range info.Rows {
		// 先处理单表
		plog.Infof("解析 %v.%v\n", tableName, subTableName)
		// 枚举替换
		if err := p.swapEnum(rows, xlsxConf.Enums); err != nil {
			plog.Error(tableName, "替换枚举出错", err)
			return
		}
		// 参数展开
		if err := p.expandParam(rows); err != nil {
			plog.Error(tableName, "参数展开出错", err)
			return
		}
		// ...其他展开

		// 创建数据结构, 赋值
		data, err := p.createStruct(rows)
		if err == nil {
			p.mergeMap(totalStructs, data)
		}
	}
	p.Output[tableName] = totalStructs
}

// 枚举替换
func (p *Parser) swapEnum(origin [][]string, enumItems []conf.EnumItem) error {
	for _, enumItem := range enumItems {
		if enumItem.Table == "enum" {
			tokens := strings.Split(enumItem.Sheet, ",")
			swapTableList := make([]map[string]string, 0, len(tokens))
			for _, sheet := range tokens {
				swapTable := p.EnumSwapDict[sheet]
				if swapTable == nil {
					plog.Errorf("enum子表%v不存在!!\n", sheet)
					return cmn.ErrNotExist
				}
				swapTableList = append(swapTableList, swapTable)
			}
			if err := p.swapEnumFieldMultiTable(origin, enumItem.Field, swapTableList); err != nil {
				return err
			}
		} else {
			tokens := strings.Split(enumItem.Table, ",")
			swapTableList := make([]map[string]string, 0, len(tokens))
			for _, table := range tokens {
				// 其他表单替换
				xlsxInfo := p.Xlsx[table]
				if xlsxInfo == nil || xlsxInfo.NameDict == nil {
					plog.Errorf("xlsx表%v不存在!!不存在name->id键值对!!\n", table)
					return cmn.ErrNotExist
				}
				swapTableList = append(swapTableList, xlsxInfo.NameDict)
			}
			if err := p.swapEnumFieldMultiTable(origin, enumItem.Field, swapTableList); err != nil {
				return err
			}
		}
	}
	return nil
}

// 替换枚举列，把数据的field列内容进行替换
func (p *Parser) swapEnumField(origin [][]string, field string, swapTable map[string]string) error {
	var ok bool
	var err error = nil
	var newValue string

	column := 0
	for ; column < len(origin[0]); column++ {
		if origin[1][column] != field {
			continue
		}

		for rowIdx := 0; rowIdx < len(origin); rowIdx++ {
			// 前ignoreLine行是结构，不替换
			if rowIdx < conf.Cfg.IgnoreLine {
				continue
			}
			if strings.ToUpper(origin[rowIdx][column]) == "NULL" {
				continue
			}
			// 原本就已经是数字了，不需要枚举
			_, e := strconv.ParseInt(origin[rowIdx][column], 10, 64)
			if e == nil {
				continue
			}

			newValue, ok = swapTable[origin[rowIdx][column]]
			if ok == false {
				plog.Errorf("枚举值%v不存在 第%d行第%d列\n", origin[rowIdx][column], rowIdx, column)
				err = cmn.ErrFail
				continue
			}
			origin[rowIdx][column] = newValue
		}
	}
	return err
}

func (p *Parser) swapEnumFieldMultiTable(origin [][]string, field string, swapTableList []map[string]string) error {
	var ok bool
	var err error = nil
	var newValue string

	for column := 0; column < len(origin[0]); column++ {
		if origin[1][column] != field {
			continue
		}

		for rowIdx := 0; rowIdx < len(origin); rowIdx++ {
			// 前ignoreLine行是结构，不替换
			if rowIdx < conf.Cfg.IgnoreLine {
				continue
			}
			// 原本就是NULL
			if strings.ToUpper(origin[rowIdx][column]) == "NULL" {
				continue
			}
			// 原本就已经是数字了，不需要枚举
			_, e := strconv.ParseInt(origin[rowIdx][column], 10, 64)
			if e == nil {
				continue
			}

			for _, swapTable := range swapTableList {
				newValue, ok = swapTable[origin[rowIdx][column]]
				if ok {
					origin[rowIdx][column] = newValue
					goto NextRowId
				}
			}

			plog.Errorf("枚举值%v不存在 第%d行第%d列\n", origin[rowIdx][column], rowIdx, column)
			err = cmn.ErrFail
		NextRowId:
		}
	}
	return err
}

// 参数展开
func (p *Parser) expandParam(rows [][]string) error {
	// todo
	return nil
}

// 多语言替换
func (p *Parser) swapLocale(origin [][]string, LocaleItems []conf.LocaleItem, locale string) error {
	for _, localeItem := range LocaleItems {
		localeSwapTable := p.LocaleSwapDict[localeItem.Table]
		if localeSwapTable == nil {
			plog.Errorf("多语言替换子表%v不存在!!\n", localeItem.Table)
			return cmn.ErrNotExist
		}
		swapTable := localeSwapTable[locale]
		if swapTable == nil {
			plog.Errorf("多语言替换表%v不存在语言%v!!\n", localeItem.Table, locale)
			return cmn.ErrNotExist
		}
		p.swapEnumField(origin, localeItem.Field, swapTable)
	}
	return nil
}

// 创建数据结构, map->struct, id作为索引
func (p *Parser) createStruct(rows [][]string) (map[int]interface{}, error) {
	data := make(map[int]interface{})
	if len(rows) < conf.Cfg.IgnoreLine {
		plog.Errorf("错误xlsx数据格式，表头只有%v行，不足%v行\n", len(rows), conf.Cfg.IgnoreLine)
		return nil, cmn.ErrNull
	}
	builder := NewStructBuilder(rows[:conf.Cfg.IgnoreLine])
	if err := builder.BuildStruct(); err != nil {
		return nil, err
	}
	//b := NewGoCodeBuilder("", "", "")
	//b.DebugType(builder.StructType, "")
	for rowIdx, row := range rows {
		if rowIdx < conf.Cfg.IgnoreLine {
			continue
		}
		id, value, err := builder.CreateInstance(row)
		if err != nil {
			plog.Errorf("解析第%d行数据失败，错误%v\n", rowIdx, err)
			continue
		}
		if data[id] != nil {
			plog.Errorf("解析第%d行数据ID重复！！ id = %d\n", rowIdx, id)
			continue
		}
		data[id] = value
	}
	return data, nil
}

func (p *Parser) outputJson() {
	// 导出数据文件
	for tableName, outputData := range p.Output {
		if conf.OutputJson(tableName) == false {
			continue
		}
		outputFile, err := os.Create(p.outputDir + "/" + tableName + ".json")
		if err != nil {
			plog.Error(tableName, "生成文件失败", err)
			continue
		}
		bs, err := json.Marshal(outputData)
		if err != nil {
			plog.Error(tableName, "json.Marshal fail!!", err)
			continue
		}
		if _, err = outputFile.Write(bs); err != nil {
			plog.Error(tableName, "fail to file.Write!!", err)
			continue
		}
		//encoder := json.NewEncoder(outputFile) 这种编码会导致输出unix文件
		//if err = encoder.Encode(outputData); err != nil {
		//	plog.Error(tableName, "导出json失败", err)
		//	continue
		//}
	}

}

// 合并map数据
func (p *Parser) mergeMap(origin map[int]interface{}, addMap map[int]interface{}) {
	for id, value := range addMap {
		if origin[id] != nil {
			plog.Error("ID冲突 id=", id)
			continue
		}
		origin[id] = value
	}
}

// 生成golang桩文件
func (p *Parser) genGolangStub(codeDir string) {
	plog.Trace()
	for fileName, data := range p.Output {
		if conf.OutputGo(fileName) == false {
			continue
		}
		for _, v := range data {
			c := NewGoCodeBuilder(codeDir, p.outputDir, fileName)
			c.GenStructWithName(v, fileName)
			c.Output()
			break
		}
	}
}

// 生成C#桩文件
func (p *Parser) genCsharpStub(codeDir string) {
	plog.Trace()
	for fileName, data := range p.Output {
		if conf.OutputCs(fileName) == false {
			continue
		}
		for _, v := range data {
			var csFileName = cmn.CamelName(fileName)
			c := NewCsharpCodeBuilder(codeDir, p.outputDir, csFileName)
			c.GenStructWithName(v, csFileName)
			c.Output()
			break
		}
	}
}
