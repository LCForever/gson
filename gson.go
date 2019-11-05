package gson

import (
	"bufio"
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
	}
}

func parseValue(reader *bufio.Reader) Value {
	escapeWhiteSpace(reader)
	item, err := reader.ReadByte()
	if nil != err {
		panic("value error in the json string")
	}
	switch item {
	case '{':
		parseObject(reader)
	case '[':
		parseArray(reader)
	case '"':
		parseString(reader)
	case 'n':
		parseNil(reader)
	case 't':
		parseTrue(reader)
	case 'f':
		parseFalse(reader)
	default:
		parseNumber(reader)
	}
}

func parse(reader *bufio.Reader) *Value {
	escapeWhiteSpace(reader)
	//the format of json is invalid
	item, err := reader.ReadByte()
	if '{' != item {
		panic("json string not started with {")
	}
	return parseObject(reader)
}

func parseObject(reader *bufio.Reader) *Value {
	escapeWhiteSpace(reader)
	res := &Value{vObjectType, unsafe.Pointer(new(JsonObject))}
	item, err := reader.ReadByte()
	if nil == err {
		panic("object invalid in the json string")
	}
	//empty objext
	if '}' == item {
		return res
	}
	for {
		//will take the next item
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil != err || '"' != item {
			panic("json key is not a string.")
		}
		key := parseString(reader)
		pStrKey := (*string)(key.ptrValue)
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil != err || ':' != item {
			panic("invalid json format")
		}
		val := parseValue(reader)
		((*JsonObject)(res.ptrValue)).lstObjects = append(((*JsonObject)(res.ptrValue)).lstObjects, Member{pStrKey, val})
		//another member, do not take next item first
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil == err {
			panic("invalid json format")
		}
		if ',' == item {
			continue
		}
		if '}' == item {
			break
		}
		panic("invalid json format")
	}
	return res
}

func parseArray(reader *bufio.Reader) *Value {
	escapeWhiteSpace(reader)
	item, err := reader.ReadByte()
	if nil == err {
		panic("array invalid in the json string")
	}
	res := &Value{vArrayType, unsafe.Pointer(new(JsonArray))}
	//empty array
	if ']' == item {
		return res
	}
	for {
		//will take the next item
		val := parseValue(reader)
		((*JsonArray)(res.ptrValue)).lstValues = append(((*JsonArray)(res.ptrValue)).lstValues, val)
		escapeWhiteSpace(reader)
		item, err = reader.ReadByte()
		if nil == err {
			panic("invalid json format")
		}
		//another member, do not take next item first
		if ',' == item {
			continue
		}
		if ']' == item {
			break
		}
		panic("invalid json format")
	}
	return res
}

func parseString(reader *bufio.Reader) *Value {

}

func parseNil(reader *bufio.Reader) *Value {
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

}
