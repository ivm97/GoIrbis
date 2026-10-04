package irbis

import (
	"encoding/binary"
	"io"
	"strings"
	"unicode"
)

const FullDelimiter = "\x1F\x1E"
const FirstDelimiter = "\x1F"
const SecondDelimiter = "\x1E"

func contains(s []int, item int) bool {
	for _, one := range s {
		if one == item {
			return true
		}
	}

	return false
}

func ToAnsi(text string) []byte {
	return cp1251FromUnicode(text)
}

func FromAnsi(buffer []byte) string {
	return cp1251ToUnicode(buffer)
}

func toUtf8(text string) []byte {
	result := []byte(text)
	return result
}

func fromUtf8(buffer []byte) string {
	return string(buffer)
}

func removeComments(text string) string {
	if len(text) == 0 || !strings.Contains(text, "/*") {
		return text
	}

	result := strings.Builder{}
	state := '\x00'
	chars := []rune(text)
	index := 0
	length := len(chars)
	result.Grow(length)

	for index < length {
		c := chars[index]

		switch state {
		case '\'', '"', '|':
			if c == state {
				state = '\x00'
			}
			result.WriteRune(c)

		default:
			if c == '/' {
				if (index+1 < length) && (chars[index+1] == '*') {
					for index < length {
						c = chars[index]
						if (c == '\r') || (c == '\n') {
							result.WriteRune(c)
							break
						}
						index++
					}
				} else {
					result.WriteRune(c)
				}
			} else if (c == '\'') || (c == '"') || (c == '|') {
				state = c
				result.WriteRune(c)
			} else {
				result.WriteRune(c)
			}
		}

		index++
	}

	return result.String()
}

func prepareFormat(text string) string {
	text = removeComments(text)
	length := len(text)
	if length == 0 {
		return text
	}

	flag := false
	chars := []rune(text)
	for i := range chars {
		if chars[i] < ' ' {
			flag = true
			break
		}
	}

	if !flag {
		return text
	}

	result := strings.Builder{}
	result.Grow(length)
	for i := range chars {
		c := chars[i]
		if c >= ' ' {
			result.WriteRune(c)
		}
	}

	return result.String()
}

func DosToIrbis(text string) string {
	return strings.ReplaceAll(text, "\n", FullDelimiter)
}

func IrbisToDos(text string) string {
	return strings.ReplaceAll(text, FullDelimiter, "\n")
}

func IrbisToLines(text string) []string {
	return strings.Split(text, FullDelimiter)
}

func LinesToIrbis(lines []string) string {
	result := strings.Builder{}
	for _, line := range lines {
		result.WriteString(line)
		result.WriteString(FullDelimiter)
	}

	return result.String()
}

func LeftPad(s string, n int) string {
	delta := n - len(s)
	if delta <= 0 {
		return s
	}

	result := strings.Builder{}
	result.Grow(n)
	for i := 0; i < delta; i++ {
		result.WriteRune(' ')
	}
	result.WriteString(s)

	return result.String()
}

func RightPad(s string, n int) string {
	delta := n - len(s)
	if delta <= 0 {
		return s
	}

	result := strings.Builder{}
	result.Grow(n)
	result.WriteString(s)
	for i := 0; i < delta; i++ {
		result.WriteRune(' ')
	}

	return result.String()
}

func PickOne(lines ...string) string {
	for _, line := range lines {
		if len(line) != 0 {
			return line
		}
	}

	return ""
}

func ParseInt32(buffer []byte) (result int) {
	for _, b := range buffer {
		result = result*10 + int(b-'0')
	}

	return
}

// ReadInt16 считывает из потока короткое целое в сетевом формате ИРБИС64.
func ReadInt16(reader io.Reader) (result int16) {
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return
}

// ReadInt32 считывает из потока целое число в сетевом формате ИРБИС64.
func ReadInt32(reader io.Reader) (result int32) {
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return
}

// ReadInt64 считывает из потока длинное целое в сетевом формате ИРБИС64.
func ReadInt64(reader io.Reader) (result int64) {
	var low, high int32
	if err := binary.Read(reader, binary.BigEndian, &low); err != nil {
		panic(err)
	}
	if err := binary.Read(reader, binary.BigEndian, &high); err != nil {
		panic(err)
	}
	result = (int64(high) << 32) + int64(low)
	return
}

func SameRune(left, right rune) bool {
	return unicode.ToUpper(left) == unicode.ToUpper(right)
}

func SameString(left, right string) bool {
	return strings.EqualFold(left, right)
}

func SplitLines(text string) []string {
	// TODO implement properly
	return strings.Split(text, "\n")
}

func trimLeft(text string) string {
	index := 0
	length := len(text)
	for index < length {
		if text[index] != ' ' {
			break
		}
		index++
	}

	return text[index:]
}

func trimRight(text string) string {
	length := len(text)
	index := length - 1
	for index >= 0 {
		if text[index] != ' ' {
			break
		}
		index--
	}

	return text[:index]
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
