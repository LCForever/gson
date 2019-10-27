package gson
//this file defines some consts and the interface will be used
//the type of the value, used in recording the value type
const(
	vObjectType = 0x00
	vArrayType = 0x01
	vBoolType = 0x02
	vUIntType = 0x04
	vIntType = 0x08
	vDoubleType = 0x10
	vStringType = 0x20
)
type Value struct{
	m_vType uint8	//the type of the value
	m_ptrValue PointerItem
}

//a member is a k-v pair
type Member struct{
	m_strKey string
	m_value Value
}
//a JsonObject is an array of members
type JsonObject struct{
	m_lstObjects []Member
}

type JsonArray struct{
	m_lstValues []Value
}