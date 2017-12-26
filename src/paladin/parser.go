package paladin

import (
	"conf"
	"encoding/json"
	"frm/plog"
	"os"
	"strings"
)

type Parser struct {
	// 数据
	Enum   *XlsxInfo
	Locale *XlsxInfo            // 多语言
	Xlsx   map[string]*XlsxInfo // 表代号->xlsx数据

	// 生成数据
	Output map[string]interface{} // 生成待输出的数据

	// 内部使用变量
	outputDir string
	genGolang bool
	genCsharp bool
}

func NewParser(outputDir string, genGolang bool, genCsharp bool) *Parser {
	p := new(Parser)
	p.Xlsx = make(map[string]*XlsxInfo)
	p.Output = make(map[string]interface{})
	p.outputDir = outputDir
	p.genGolang = genGolang
	p.genCsharp = genCsharp
	return p
}

func (p *Parser) Start() {
	// 加载文件
	p.loadFiles()

	// 解析数据文件
	p.parseData()

	// 解析多语言文件

	// 导出为json
	p.output("json")

	// 是否生成桩文件
	p.genStubCode()
}

// 加载文件
func (p *Parser) loadFiles() {
	// 加载枚举文件
	reader := NewXlsxReader(false, false)
	enumInfo, err := reader.Read("enum", conf.Cfg.EnumFile)
	if err != nil {
		plog.Panic("读取枚举表失败!!", err)
	}
	p.Enum = enumInfo

	// 加载xlsx文件
	for tblName, tbl := range conf.Cfg.Tables {
		reader := NewXlsxReader(tbl.AutoId, tbl.Horizontal)
		if err = reader.LoadEnum(tbl.Enums); err != nil {
			plog.Errorf("%s加载枚举配置失败！错误码:%v\n", tbl.Workbook, err)
		}
		info, err := reader.Read(tblName, tbl.Workbook)
		if err != nil {
			plog.Errorf("读取%s失败！错误码:%v\n", tbl.Workbook, err)
		}
		p.Xlsx[tblName] = info
	}

	// 加载多语言文件
	reader = NewXlsxReader(false, false)
	localeInfo, err := reader.Read("locale", conf.Cfg.LocaleFile)
	if err != nil {
		plog.Panic("读取多语言表失败!!", err)
	}
	p.Locale = localeInfo
}

// 解析
func (p *Parser) parseData() {
	for tableName, info := range p.Xlsx {
		p.parseXlsx(tableName, info)
	}
}

// 导出
func (p *Parser) output(fmt string) {
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "json":
		p.outputJson()

	default:
		plog.Panic("Invalid output fmt", fmt)
	}
}

// 生成桩文件
func (p *Parser) genStubCode() {
	if p.genGolang {
		p.genGolangStub()
	}
	if p.genCsharp {
		p.genCsharpStub()
	}
}

////////////////////////////////////// 子函数 //////////////////////////////////////
func (p *Parser) parseXlsx(tableName string, info *XlsxInfo) {
	plog.Trace(tableName)
	for subTableName, rows := range info.Rows {
		plog.Trace("解析", subTableName)
		// 枚举替换
		if err := p.swapEnum(info); err != nil {
			plog.Error(tableName, "替换枚举出错", err)
		}
		// 参数展开
		if err := p.expandParam(info); err != nil {
			plog.Error(tableName, "参数展开出错", err)
		}
		// ...其他展开

		// 创建数据结构, 赋值
		p.Output[tableName] = p.createStruct()
	}
}

// 枚举替换
func (p *Parser) swapEnum(info *XlsxInfo) error {
	// todo
	// rows[0] 是数据类型
	// rows[1] 是数据名称
	// rows[2] 是辅助记忆描述，可忽略
	return nil
}

// 参数展开
func (p *Parser) expandParam(info *XlsxInfo) error {
	// todo
	return nil
}

// 创建数据结构, 赋值
func (p *Parser) createStruct() interface{} {
	// rows[0] 是数据类型
	// rows[1] 是数据名称
	// rows[2] 是辅助记忆描述，可忽略
	return nil
}

func (p *Parser) outputJson() {
	for tableName, outputData := range p.Output {
		outputFile, err := os.Create(p.outputDir + tableName + ".json")
		if err != nil {
			plog.Error(tableName, "生成文件失败", err)
		}
		encoder := json.NewEncoder(outputFile)
		if err = encoder.Encode(outputData); err != nil {
			plog.Error(tableName, "导出json失败", err)
		}
	}
}
