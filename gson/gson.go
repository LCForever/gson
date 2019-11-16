package gson

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

const (
	PATH_SPLIT_CHAR = '.'
)

//dom style
type Gson struct {
	m_val *Value
}

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
func (gson *Gson) Parse(reader *bufio.Reader) (bRes bool) {
	defer func() {
		if err := recover(); nil != err {
			fmt.Println(err)
			bRes = false
		}
	}()
	bRes = true
	gson.m_val = parse(reader)
	return
}

//export Dump
func (gson *Gson) Dump() (string, bool) {
	if nil == gson.m_val {
		return "", false
	}
	return gson.m_val.Dump(), true
}

//to be different with number, path should have quoto
//for example, `"a"."b"."c".1`
func (gson *Gson) Get(strPath string) (res *Value) {
	defer func() {
		if err := recover(); nil != err {
			fmt.Println("invalid path")
			res = nil
		}
	}()
	res = gson.m_val
	reader := bufio.NewReader(strings.NewReader(strPath))
	var item byte
	var err error
	var bHas bool
	item, err = reader.ReadByte()
	//empty path
	if nil != err {
		return
	}
	if PATH_SPLIT_CHAR == item {
		panic("Invalid path")
	}

	for {
		if '"' == item || ('0' <= item && '9' >= item) {
			if '"' == item {
				if vObjectType != res.vType {
					panic("invalid path")
				}
				pResKey := parseString(reader)
				res, bHas = ((*JsonObject)(res.ptrValue)).mapObjects[*(*string)(pResKey.ptrValue)]
				if !bHas {
					panic("invalid path")
				}
			} else if '0' <= item && '9' >= item {
				if vArrayType != res.vType {
					panic("invalid path")
				}
				reader.UnreadByte()
				pathIndex := getPathIndex(reader)
				if pathIndex >= len(((*JsonArray)(res.ptrValue)).lstValues) {
					panic("invalid path")
				}
				res = ((*JsonArray)(res.ptrValue)).lstValues[pathIndex]
			}
			item, err = reader.ReadByte()
			//finish byte
			if nil != err {
				break
			}
		}

		if PATH_SPLIT_CHAR == item {
			//cannot have multiple splitor
			item, err = reader.ReadByte()
			//empty path
			if nil != err || PATH_SPLIT_CHAR == item {
				panic("Invalid path")
			}
		} else {
			panic("Invalid path")
		}
	}
	return
}

//get the original string of the specified path
func (gson *Gson) Original(strPath string) (string, error) {
	value := gson.Get(strPath)
	if nil != value {
		return value.Dump(), nil
	}
	return "", errors.New("invalid path")
}

//the origval should be a correct value
func getValueByString(strOrigVal string) *Value {
	reader := bufio.NewReader(strings.NewReader(strOrigVal))
	newVal := parseValue(reader)
	escapeWhiteSpace(reader)
	//should be finished
	_, err := reader.ReadByte()
	if nil != newVal && nil != err {
		return newVal
	}
	return nil
}

//Set item, need to have the key with the path
func (gson *Gson) Set(strPath, strOrigVal string) (resErr error) {
	defer func() {
		if nil != recover() {
			resErr = errors.New("Invalid value format")
		}
	}()
	value := gson.Get(strPath)
	if nil != value {
		newVal := getValueByString(strOrigVal)
		if nil != newVal {
			*value = *newVal
			return nil
		}
	} else {
		return errors.New("Invalid path")
	}
	return errors.New("Invalid value format")
}

//AddObject should have path, the item related with the path should be an object
//There should also be key and value in the function
func (gson *Gson) AddObject(strPath, strKey, strOrigVal string) (resErr error) {
	defer func() {
		if nil != recover() {
			resErr = errors.New("Invalid value format")
		}
	}()
	value := gson.Get(strPath)
	if nil == value {
		return errors.New("Invalid path")
	}
	if vObjectType != value.vType {
		return errors.New("The item with specified path should be an object.")
	}
	_, ok := ((*JsonObject)(value.ptrValue)).mapObjects[strKey]
	if ok {
		return errors.New("The key already exist.")
	}
	newVal := getValueByString(strOrigVal)
	if nil == newVal {
		return errors.New("Invalid value format")
	}
	((*JsonObject)(value.ptrValue)).mapObjects[strKey] = newVal
	return nil
}

//AddMember should have path, the item related with the path should be an array
//There should also be a value in the function
func (gson *Gson) AddMember(strPath, strOrigVal string) (resErr error) {
	defer func() {
		if nil != recover() {
			resErr = errors.New("Invalid value format")
		}
	}()
	value := gson.Get(strPath)
	if nil == value {
		return errors.New("Invalid path")
	}
	if vArrayType != value.vType {
		return errors.New("The item with specified path should be an array.")
	}
	newVal := getValueByString(strOrigVal)
	if nil == newVal {
		return errors.New("Invalid value format")
	}
	((*JsonArray)(value.ptrValue)).lstValues = append(((*JsonArray)(value.ptrValue)).lstValues, newVal)
	return nil
}

func getPathIndex(reader *bufio.Reader) int {
	var resIndex int = 0
	var bHasValue bool = false
	for {
		item, err := reader.ReadByte()
		if nil != err {
			if true == bHasValue {
				return resIndex
			}
			panic("Invalid path index")
		}
		if '0' <= item && '9' >= item {
			resIndex = 10*resIndex + (int)(item-'0')
			bHasValue = true
		} else if PATH_SPLIT_CHAR == item {
			reader.UnreadByte()
			break
		} else {
			panic("Invalid path index")
		}
	}
	return resIndex
}

func parse(reader *bufio.Reader) *Value {
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
	escapeWhiteSpace(reader)
	res := &Value{vObjectType, unsafe.Pointer(new(JsonObject))}
	((*JsonObject)(res.ptrValue)).mapObjects = map[string]*Value{}
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
		_, ok := ((*JsonObject)(res.ptrValue)).mapObjects[*pStrKey]
		if ok {
			panic("duplicate key in json object")
		}
		((*JsonObject)(res.ptrValue)).mapObjects[*pStrKey] = parseValue(reader)
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
	res := &Value{}
	res.vType = vStringType
	byteRes := make([]byte, 0)
	byteSpecial := map[byte]byte{'b': 8, 'f': 12, 'n': 10, 'r': 13, 't': 9, '\\': 92, '"': 34}
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
	res.vType = vFalseType
	reader.Discard(4)
	return res
}

func parseNumber(reader *bufio.Reader) *Value {
	var item byte
	var byteNum []byte
	var err error
	for {
		item, err = reader.ReadByte()
		//read end, just parse the number, no need to deal here, will be dealt in parseArray or parseObject
		if nil != err {
			break
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
		return &Value{vNumType, unsafe.Pointer(&Number{vIntType, unsafe.Pointer(&resNum), &strNum})}
	}
	//maybe number exceed int range
	if resNum, err := strconv.ParseUint(strNum, 10, 64); nil == err {
		return &Value{vNumType, unsafe.Pointer(&Number{vUIntType, unsafe.Pointer(&resNum), &strNum})}
	}
	if resNum, err := strconv.ParseFloat(strNum, 64); nil == err {
		return &Value{vNumType, unsafe.Pointer(&Number{vDoubleType, unsafe.Pointer(&resNum), &strNum})}
	}
	panic("invalid json number")
}
