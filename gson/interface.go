package gson

import (
	"bytes"
	"errors"
	"math"
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

func (value *Value) GetBoolValue() (bool, error) {
	if vTrueType == value.vType {
		return true, nil
	}
	if vFalseType == value.vType {
		return false, nil
	}
	return false, errors.New("value not bool type")
}

func (value *Value) GetIntValue() (int64, error) {
	if value.IsNumber() {
		numValue := (*Number)(value.ptrValue)
		if vIntType == numValue.vType {
			return *(*int64)(numValue.ptrValue), nil
		}
		if vUIntType == numValue.vType {
			unRes := *(*uint64)(numValue.ptrValue)
			if math.MaxInt64 >= unRes {
				return (int64)(unRes), nil
			}
		}
	}
	return 0, errors.New("value not int64 type")
}

func (value *Value) GetUIntValue() (uint64, error) {
	if value.IsNumber() {
		numValue := (*Number)(value.ptrValue)
		if vUIntType == numValue.vType {
			return *(*uint64)(numValue.ptrValue), nil
		}
		if vIntType == numValue.vType {
			nRes := *(*int64)(numValue.ptrValue)
			if 0 <= nRes {
				return (uint64)(nRes), nil
			}
		}
	}
	return 0, errors.New("value not uint64 type")
}

func (value *Value) GetDoubleValue() (float64, error) {
	if value.IsNumber() {
		numValue := (*Number)(value.ptrValue)
		if vUIntType == numValue.vType {
			return (float64)(*(*uint64)(numValue.ptrValue)), nil
		}
		if vIntType == numValue.vType {
			return (float64)(*(*int64)(numValue.ptrValue)), nil
		}
		if vDoubleType == numValue.vType {
			return (*(*float64)(numValue.ptrValue)), nil
		}
	}
	return 0, errors.New("value not float type")
}

func (value *Value) GetStringValue() (string, error) {
	if value.IsString() {
		return (*(*string)(value.ptrValue)), nil
	}
	return "", errors.New("value not string type")
}

func (value *Value) IsObject() bool {
	return vObjectType == value.vType
}

func (value *Value) IsArray() bool {
	return vArrayType == value.vType
}

func (value *Value) IsNil() bool {
	return vNilType == value.vType
}

func (value *Value) IsNumber() bool {
	return vNumType == value.vType
}

func (value *Value) IsString() bool {
	return vStringType == value.vType
}

func (value *Value) IsBool() bool {
	return vTrueType == value.vType || vFalseType == value.vType
}

func (value *Value) Dump() string {
	switch value.vType {
	case vObjectType:
		return ((*JsonObject)(value.ptrValue)).Dump()
	case vArrayType:
		return ((*JsonArray)(value.ptrValue)).Dump()
	case vTrueType:
		return "true"
	case vFalseType:
		return "false"
	case vNilType:
		return "null"
	case vNumType:
		return ((*Number)(value.ptrValue)).Dump()
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

func (num *Number) Dump() string {
	return *num.pStrRaw
}

func (jsonObject *JsonObject) Dump() string {
	strRes := "{"
	nLen := len(jsonObject.mapObjects)
	nIndex := 0
	for k, v := range jsonObject.mapObjects {
		strRes += `"` + dumpString(k) + `":` + v.Dump()
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

func (jsonArray *JsonArray) Dump() string {
	strRes := "["
	nLen := len(jsonArray.lstValues)
	for nIndex, item := range jsonArray.lstValues {
		strRes += item.Dump()
		if nLen-1 > nIndex {
			strRes += ","
		}
	}
	strRes += "]"
	return strRes
}
