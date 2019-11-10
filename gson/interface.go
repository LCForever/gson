package gson

import (
	"bytes"
	"strconv"
	"unsafe"
)

//this file defines some consts and the interface will be used
//the type of the value, used in recording the value type
const (
	vObjectType = 0x00
	vArrayType  = 0x01
	vTrueType   = 0x02
	vFalseType  = 0x04
	vNilType    = 0x08
	vUIntType   = 0x10
	vIntType    = 0x20
	vDoubleType = 0x40
	vStringType = 0x80
)

type Value struct {
	vType    uint8 //the type of the value
	ptrValue unsafe.Pointer
}

func (value *Value) dump() string {
	switch value.vType {
	case vObjectType:
		return ((*JsonObject)(value.ptrValue)).dump()
	case vArrayType:
		return ((*JsonArray)(value.ptrValue)).dump()
	case vTrueType:
		return "true"
	case vFalseType:
		return "false"
	case vNilType:
		return "null"
	case vUIntType:
		return strconv.FormatUint(*((*uint64)(value.ptrValue)), 10)
	case vIntType:
		return strconv.FormatInt(*((*int64)(value.ptrValue)), 10)
	case vDoubleType:
		return strconv.FormatFloat(*((*float64)(value.ptrValue)), 'e', -1, 64)
	}
	return `"` + dumpString(*(*string)(value.ptrValue)) + `"`
}

func dumpString(strValue string) string {
	var strBuf bytes.Buffer
	byteSpecial := map[byte]string{'\b': `\b`, '\f': `\f`, '\n': `\n`, '\r': `\r`, '\t': `\t`, '\\': `\\`, '"': `\"`}
	for nIndex := 0; nIndex < len(strValue); nIndex++ {
		value, ok := byteSpecial[strValue[nIndex]]
		if ok {
			strBuf.WriteString(value)
			continue
		}
		strBuf.WriteByte(strValue[nIndex])
	}
	return strBuf.String()
}

//a member is a k-v pair
type Member struct {
	pStrKey *string
	value   *Value
}

func (member *Member) dump() string {
	return `"` + dumpString(*(member.pStrKey)) + `":` + member.value.dump()
}

//JsonObject is an array of members
type JsonObject struct {
	lstObjects []*Member
}

func (jsonObject *JsonObject) dump() string {
	strRes := "{"
	nLen := len(jsonObject.lstObjects)
	for nIndex, item := range jsonObject.lstObjects {
		strRes += item.dump()
		if nLen-1 > nIndex {
			strRes += ","
		}
	}
	strRes += "}"
	return strRes
}

type JsonArray struct {
	lstValues []*Value
}

func (jsonArray *JsonArray) dump() string {
	strRes := "["
	nLen := len(jsonArray.lstValues)
	for nIndex, item := range jsonArray.lstValues {
		strRes += item.dump()
		if nLen-1 > nIndex {
			strRes += ","
		}
	}
	strRes += "]"
	return strRes
}
