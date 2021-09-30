package utils

import (
	"fmt"
	"strings"
)

func LowerFirstLetter(str string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(str[0:1]), str[1:])
}

func UpperFirstLetter(str string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(str[0:1]), str[1:])
}
