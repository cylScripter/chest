package utils

import (
	"fmt"
	"strings"
	"unicode"
)

func EscapeMysqlString(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		escape = 0
		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
		case '\n': /* Must be escaped for logs */
			escape = 'n'
		case '\r':
			escape = 'r'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"': /* Better safe than sorry */
			escape = '"'
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}
		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}
	return string(dest)
}

// CamelToSnake 将驼峰命名法转换为下划线命名法
func CamelToSnake(modelType string) string {
	var modelPrefix string
	var s string
	sList := strings.Split(modelType, ".")
	fmt.Println(sList)
	if len(sList) >= 2 {
		modelPrefix = sList[0]
		s = sList[1]
	} else {
		s = modelType
	}
	var newStr string

	fmt.Println(s)
	if len(s) > 5 && s[0:5] == "Model" {
		newStr = s[5:]
	} else {
		newStr = s
	}
	var result []rune
	for i, r := range newStr {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}

	if modelPrefix != "" {
		modelPrefix = strings.ToLower(modelPrefix)
		return modelPrefix + "_" + string(result)
	}
	return string(result)
}
