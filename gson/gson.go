package gson

import (
	"bufio"
	"fmt"
	"strconv"
	"unsafe"
)

/*
//dom style
type gson struct {
	m_obj JsonObject
}
*/

func isWhiteSpace(item byte) bool {
	return ' ' == item || '\t' == item || '\n' == item
}

func isSplitByte(item byte) bool {
	return isWhiteSpace(item) || ',' == item || ']' == item || '}' == item
}

func escapeWhiteSpace(reader *bufio.Reader) {
	for {
		item, err := reader.ReadByte()
		if nil == err && isWhiteSpace(item) {
			continue
		} else if nil != err {
			break
		}
		//not white space, push back the byte
		reader.UnreadByte()
		break
	}
}

func parseValue(reader *bufio.Reader) *Value {
	fmt.Println("parseValue")
	escapeWhiteSpace(reader)
	item, err := reader.ReadByte()
	if nil != err {
		panic("json string not finished")
	}
	switch item {
	case '{':
		return parseObject(reader)
	case '[':
		return parseArray(reader)
	case '"':
		return parseString(reader)
	case 'n':
		return parseNil(reader)
	case 't':
		return parseTrue(reader)
	case 'f':
		return parseFalse(reader)
	}
	//we need to use the byte, parse as number
	reader.UnreadByte()
	return parseNumber(reader)
}

//export Parse
func Parse(reader *bufio.Reader) *Value {
	fmt.Println("Parse")
	escapeWhiteSpace(reader)
	//the format of json is invalid
	item, err := reader.ReadByte()
	if nil != err {
		panic("invalid json format: no start")
	}
	if '{' != item {
		panic("json string not started with {")
	}
	value := parseObject(reader)
	escapeWhiteSpace(reader)
	item, err = reader.ReadByte()
	if nil == err {
		panic("invalid json string: " + string(item))
	}
	return value
}

func parseObject(reader *bufio.Reader) *Value {
	fmt.Println("parseObject")
	escapeWhiteSpace(reader)
	res := &Value{vObjectType, unsafe.Pointer(new(JsonObject))}
	item, err := reader.ReadByte()
	if nil != err {
		panic("json string not finished")
	}
	//empty objext
	if '}' == item {
		return res
	}
	for {
		if nil != err || '"' != item {
			panic("json key is not a string.")
		}
		key := parseString(reader)
		pStrKey := (*string)(key.ptrValue)
		if 0 == len(*pStrKey) {
			panic("json key cannot be empty.")
		}
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil != err || ':' != item {
			panic("invalid json format, no value part")
		}
		val := parseValue(reader)
		((*JsonObject)(res.ptrValue)).lstObjects = append(((*JsonObject)(res.ptrValue)).lstObjects, &Member{pStrKey, val})
		//another member, do not take next item first
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil != err {
			panic("json string not finished")
		}
		if ',' == item {
			//will take the next item
			escapeWhiteSpace(reader)
			item, err = reader.ReadByte()
			continue
		}
		if '}' == item {
			break
		}
		panic("invalid json format: " + string(item))
	}
	return res
}

func parseArray(reader *bufio.Reader) *Value {
	fmt.Println("parseArray")
	escapeWhiteSpace(reader)
	item, err := reader.ReadByte()
	if nil != err {
		panic("array invalid in the json string, json string not finished")
	}
	res := &Value{vArrayType, unsafe.Pointer(new(JsonArray))}
	//empty array
	if ']' == item {
		return res
	}
	reader.UnreadByte()
	for {
		//will take the next item
		val := parseValue(reader)
		((*JsonArray)(res.ptrValue)).lstValues = append(((*JsonArray)(res.ptrValue)).lstValues, val)
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil != err {
			panic("json string not finished")
		}
		//another member, do not take next item first
		if ',' == item {
			continue
		}
		if ']' == item {
			break
		}
		panic("invalid json array: " + string(item))
	}
	return res
}

func parseString(reader *bufio.Reader) *Value {
	fmt.Println("parseString")
	res := &Value{}
	res.vType = vStringType
	byteRes := make([]byte, 20)
	byteSpecial := map[byte]byte{'b': 8, 'f': 12, 'n': 10, 'r': 13, 't': 9, '\\': 92, '"': 4}
	var item, nextItem byte
	var strValue string
	var err error
	for {
		item, err = reader.ReadByte()
		if nil != err {
			panic("string invalid in the json string")
		}
		if '"' == item {
			strValue = string(byteRes)
			res.ptrValue = unsafe.Pointer(&strValue)
			break
		}
		if '\\' == item {
			nextItem, err = reader.ReadByte()
			if nil != err {
				panic("string invalid in the json string")
			}
			value, ok := byteSpecial[nextItem]
			if ok {
				byteRes = append(byteRes, value)
				continue
			} else {
				panic("string invalid in the json string")
			}
		}
		byteRes = append(byteRes, item)
	}
	return res
}

func parseNil(reader *bufio.Reader) *Value {
	fmt.Println("parseNil")
	items, err := reader.Peek(3)
	if nil != err || !('u' == items[0] && 'l' == items[1] && 'l' == items[2]) {
		panic("invalid value")
	}
	res := &Value{}
	res.vType = vNilType
	reader.Discard(3)
	return res
}

func parseTrue(reader *bufio.Reader) *Value {
	fmt.Println("parseTrue")
	items, err := reader.Peek(3)
	if nil != err || !('r' == items[0] && 'u' == items[1] && 'e' == items[2]) {
		panic("invalid value")
	}
	res := &Value{}
	res.vType = vTrueType
	reader.Discard(3)
	return res
}

func parseFalse(reader *bufio.Reader) *Value {
	fmt.Println("parseFalse")
	items, err := reader.Peek(4)
	if nil != err || !('a' == items[0] && 'l' == items[1] && 's' == items[2] && 'e' == items[3]) {
		panic("invalid value")
	}
	res := &Value{}
	res.vType = vTrueType
	reader.Discard(4)
	return res
}

func parseNumber(reader *bufio.Reader) *Value {
	fmt.Println("parseNumber")
	var item byte
	var byteNum []byte
	var err error
	for {
		item, err = reader.ReadByte()
		if nil != err {
			panic("invalid json format")
		}
		if isSplitByte(item) {
			reader.UnreadByte()
			break
		}
		byteNum = append(byteNum, item)
	}
	var strNum string = string(byteNum)
	//parse with int firstly
	if resNum, err := strconv.ParseInt(strNum, 10, 64); nil == err {
		return &Value{vIntType, unsafe.Pointer(&resNum)}
	}
	//maybe number exceed int range
	if resNum, err := strconv.ParseUint(strNum, 10, 64); nil == err {
		return &Value{vUIntType, unsafe.Pointer(&resNum)}
	}
	if resNum, err := strconv.ParseFloat(strNum, 64); nil == err {
		return &Value{vDoubleType, unsafe.Pointer(&resNum)}
	}
	panic("invalid json number")
}
