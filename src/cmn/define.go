package cmn

import "errors"

var (
	ErrNotExist = errors.New("not exist")
	ErrFail     = errors.New("fail")
	ErrEOF      = errors.New("EOF")           // 代表结束
	ErrSkip     = errors.New("SKIP")          // 读取slice、map数据时跳过去
	ErrNull     = errors.New("null")          // 空数值
	ErrBadXlsx  = errors.New("bad xlsx data") // 错误xlsx数值
)
