package main

import (
	"os"
	"time"

	"frm/plog"

	"conf"
	"paladin"

	"gopkg.in/urfave/cli.v1"
)

// 参数解析
func main() {
	app := cli.NewApp()
	app.Name = "test"
	app.Usage = "unify test suit"
	app.Version = "0.4.7"
	app.Compiled = time.Now()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config,c",
			Usage: "config file",
			Value: "config.toml",
		},
		cli.StringFlag{
			Name:  "output,o",
			Usage: "output directory",
			Value: "data",
		},
		cli.StringFlag{
			Name:  "stub,s",
			Usage: "generated stub code files directory",
			Value: "stub",
		},
		cli.StringFlag{
			Name:  "locale,l",
			Usage: "generated locale files directory",
			Value: "locale",
		},
		cli.BoolFlag{
			Name:  "golang,go",
			Usage: "generate golang stub code",
		},
		cli.BoolFlag{
			Name:  "csharp,cs",
			Usage: "generate csharp stub code",
		},
	}
	app.Action = func(c *cli.Context) error {
		return actualMain(c)
	}
	app.Run(os.Args)
}

// 实际main函数
func actualMain(c *cli.Context) error {
	// 加载配置
	conf.Init(c.String("config"))

	// 加载log
	plog.InitLog("app.log", plog.LOG_TRACE)

	// 启动解析器
	parser := paladin.NewParser(c.String("output"), c.String("stub"), c.String("locale"), c.Bool("golang"), c.Bool("csharp"))
	parser.Start()
	return nil
}
