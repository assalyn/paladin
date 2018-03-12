package conf

import (
	"frm/plog"

	"fmt"

	"github.com/BurntSushi/toml"
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
	Workbook   string       // 工作簿
	Sheet      []string     // 子表
	AutoId     bool         // 自动id
	Horizontal bool         // 是否水平解析
	Singleton  bool         // 是否单独结构. 一般来说，只有全局配置是不需要map索引的
	Enums      []EnumItem   // 枚举替换
	Locales    []LocaleItem // 多语言替换
}

type Config struct {
	Locale     string // 多语言描述
	EnumFile   string // 枚举文件
	LocaleFile string // 多语言文件
	IgnoreLine int    // 忽略的头几行
	Tables     map[string]*Table
}

func Init(confFile string) {
	Cfg = new(Config)
	_, err := toml.DecodeFile(confFile, Cfg)
	if err != nil {
		plog.Panic(err)
	}
	//show()
}

func show() {
	fmt.Printf("%+v\n", Cfg)
	for tblName, tbl := range Cfg.Tables {
		fmt.Printf("%s %+v\n", tblName, tbl)
	}
}
