package main

import (
	"frm/plog"

	"github.com/dave/jennifer/jen"
	"github.com/tealeg/xlsx"
)

func main() {
	// 读取数据
	data := readStruct()

	// 生成数据定义文件
	gtor := NewGenerator("equip", data[:4])
	gtor.Parse()
	gtor.Write()

	//// 解析rows, 生成struct
	//f := NewFile("main")
	//f.Func().Id("main").Params().Block(
	//	Qual("fmt", "Println").Call(Lit("Hello, world")),
	//)
	//if err := f.Save("main.dbc.go"); err != nil {
	//	plog.Error(err)
	//}
}

func readStruct() [][]string {
	xlFile, err := xlsx.OpenFile("bin/xlsx/位置表.xlsx")
	if err != nil {
		plog.Errorf("fail to read %s!! %v\n", "bin/xlsx/位置表.xlsx", err)
		return nil
	}

	for _, sheet := range xlFile.Sheets {
		l := make([][]string, len(sheet.Rows))
		for rowIdx, row := range sheet.Rows {
			l[rowIdx] = make([]string, len(row.Cells))
			for column, cell := range row.Cells {
				l[rowIdx][column] = cell.Value
			}
		}
		return l
	}
	return nil
}

type Generator struct {
	*jen.File
	fileName string
	rows     [][]string
}

func NewGenerator(fileName string, data [][]string) *Generator {
	p := new(Generator)
	p.fileName = fileName + ".dbc.go"
	p.File = jen.NewFile(p.fileName)
	p.rows = data
	return p
}

// 进行解析. 能根据数据结构生成类型定义文件不？
func (p *Generator) Parse() {
	// rows[0] 是数据结构 INT
	// rows[1] 是变量名称
	// rows[2] 是数据归属
	for column := 0; column < len(p.rows[0]); column++ {

	}
}

// 导出到文件
func (p *Generator) Write() {
	p.Save(p.fileName)
}
