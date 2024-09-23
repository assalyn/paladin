package cmn

import (
	"assalyn/paladin/frm/plog"
	"regexp"
	"strings"
)

// 首字母大写，蛇形转驼峰
func CamelName(textStr string) string {
	if len(textStr) == 0 {
		return ""
	}

	// 过滤掉#
	rex, _ := regexp.Compile("#.*$")
	text := []byte(rex.ReplaceAllString(textStr, ""))
	// 字符集校验
	match, _ := regexp.Match("^[a-zA-Z][a-zA-Z0-9_]*$", text)
	if match == false {
		plog.Error("类型名含有非法字符！", string(text))
		return ""
	}

	// 首字母大写
	camel := make([]byte, 0, len(text))
	// 字符转换
	captalChar := true
	for i := 0; i < len(text); i++ {
		if text[i] == '_' {
			captalChar = true
			continue
		} else {
			if captalChar {
				camel = append(camel, ByteToUpper(text[i]))
				captalChar = false
			} else {
				camel = append(camel, text[i])
			}
		}
	}
	return string(camel)
}

func ByteToUpper(b byte) byte {
	if 'a' <= b && b <= 'z' {
		return b - 32
	}
	return b
}

// FirstCharLower 字符串首字母小写
func FirstCharLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
