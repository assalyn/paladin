package main

import (
	"assalyn/paladin/conf"
	"assalyn/paladin/frm/plog"
	"assalyn/paladin/paladin"
	"flag"
)

var (
	confFile      = flag.String("config", "config.toml", "-config config.toml")
	output        = flag.String("output", "data", "output directory")
	stub          = flag.String("stub", "stub", "generated stub code files directory")
	locale        = flag.String("locale", "locale", "generated locale files directory")
	exportJsonCfg = flag.Bool("export_json_config", false, "export config.toml in json datatype")
	golang        = flag.Bool("go", false, "generate golang stub code")
	csharp        = flag.Bool("cs", false, "generate csharp stub code")
)

// 参数解析
func main() {
	flag.Parse()

	// 加载配置
	conf.Init(*confFile)
	if exportJsonCfg != nil && *exportJsonCfg {
		conf.ExportJson("config.json")
	}

	// 加载log
	plog.InitLog("app.log", plog.LOG_TRACE)

	// 启动解析器
	parser := paladin.NewParser(*output, *stub, *locale, *golang, *csharp)
	parser.Start()
}
