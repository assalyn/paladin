package paladin

// 桩文件生成规则

import "frm/plog"

// 生成golang桩文件
func (p *Parser) genGolangStub(dir string) {
	plog.Trace()
	for fileName, data := range p.Output {
		for _, v := range data {
			c := NewCodeBuilder(fileName)
			c.GenStructWithName(v, fileName)
			c.Output(dir)
			break
		}
	}
}

// 生成C#桩文件
func (p *Parser) genCsharpStub(dir string) {
	plog.Trace()
}
