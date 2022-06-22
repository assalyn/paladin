package conf

import (
	"assalyn/paladin/frm/plog"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"strings"
)

var Cfg *Config

type EnumItem struct {
	Field string
	Table string // 表名
	Sheet string // 子表
}

type LocaleItem struct {
	Field string
	Table string
}

type Table struct {
	Workbook   string     // 工作簿
	Sheet      []string   // 子表
	AutoId     bool       // 自动id todo 未实现
	Horizontal bool       // 是否水平解析
	Output     []string   // 输出类型选项 json, cs, go. 默认全输出
	Type       string     // 表类型 server_only; client_local_read
	Enums      []EnumItem // 枚举替换
}

type Config struct {
	Locale           string // 多语言描述
	CompressSliceMap bool   // 压缩slice/map. 将只有1个字段的slice/map 内联结构展开，成为基本类型slice/map
	EnumFile         string // 枚举文件
	LocaleFile       string // 多语言文件
	IgnoreLine       int    // 忽略的头几行
	Tables           map[string]*Table
}

func Init(confFile string) {
	Cfg = new(Config)
	_, err := toml.DecodeFile(confFile, Cfg)
	if err != nil {
		plog.Panic(err)
	}
	//show()
}

func ExportJson(filename string) {
	bs, err := json.Marshal(Cfg)
	if err != nil {
		panic("invalid config file, export config.json failed!! " + err.Error())
	}
	f, err := os.Create(filename)
	if err != nil {
		panic("fail to create config.json!!" + err.Error())
	}
	if _, err = f.Write(bs); err != nil {
		panic("fail to write config to json file!!" + err.Error())
	}
}

var (
	OutputJson = outputDelegate("json")
	OutputGo   = outputDelegate("go")
	OutputCs   = outputDelegate("cs")
)

func outputDelegate(express string) func(tableName string) bool {
	return func(tableName string) bool {
		tbl := Cfg.Tables[tableName]
		if tbl == nil {
			return false
		}
		if len(tbl.Output) == 0 {
			return true
		}
		for _, regx := range tbl.Output {
			if strings.ToLower(regx) == express {
				return true
			}
		}
		return false
	}
}

func show() {
	fmt.Printf("%+v\n", Cfg)
	for tblName, tbl := range Cfg.Tables {
		fmt.Printf("%s %+v\n", tblName, tbl)
	}
}
