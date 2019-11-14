package gson

import (
	"bytes"
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
	vNumType    = 0x10
	vStringType = 0x20
)

const (
	vUIntType   = 0x00
	vIntType    = 0x01
	vDoubleType = 0x02
)

type Number struct {
	vType    uint8 //the number type of the number, vUIntType/vIntType/vDoubleType
	ptrValue unsafe.Pointer
	pStrRaw  *string //the raw string of the number
}

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
	case vNumType:
		return ((*Number)(value.ptrValue)).dump()
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

//JsonObject is an array of members
type JsonObject struct {
	mapObjects map[string]*Value
}

func (num *Number) dump() string {
	return *num.pStrRaw
}

func (jsonObject *JsonObject) dump() string {
	strRes := "{"
	nLen := len(jsonObject.mapObjects)
	nIndex := 0
	for k, v := range jsonObject.mapObjects {
		strRes += `"` + dumpString(k) + `":` + v.dump()
		if nLen-1 > nIndex {
			strRes += ","
			nIndex++
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
