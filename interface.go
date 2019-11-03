package gson

import (
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

//a member is a k-v pair
type Member struct {
	pStrKey *string
	value   Value
}

//a JsonObject is an array of members
type JsonObject struct {
	lstObjects []Member
}

type JsonArray struct {
	lstValues []Value
}
