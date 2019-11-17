package gson

//what we want to export is Value and Gson, No other structs
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
func (gson *Gson) Parse(reader *bufio.Reader) (bRes bool, errRes error) {
	defer func() {
		if err := recover(); nil != err {
			errRes = fmt.Errorf("%v", err)
			bRes = false
		}
	}()
	gson.m_val = parse(reader)
	return true, nil
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
func (gson *Gson) Get(strPath string) (res *Value, resErr error) {
	defer func() {
		if err := recover(); nil != err {
			resErr = fmt.Errorf("%v", err)
			res = nil
		}
	}()
	resErr = nil
	res = gson.m_val
	reader := bufio.NewReader(strings.NewReader(strPath))
	var item byte
	var err error
	var bHas bool
	item, err = reader.ReadByte()
	//empty path
	if nil != err {
		return res, nil
	}
	if PATH_SPLIT_CHAR == item {
		return nil, errors.New("path should not start with '.'")
	}

	for {
		if '"' == item || ('0' <= item && '9' >= item) {
			if '"' == item {
				if vObjectType != res.vType {
					return nil, errors.New("path is not match with type, expect object")
				}
				pResKey := parseString(reader)
				res, bHas = ((*tJsonObject)(res.ptrValue)).mapObjects[*(*string)(pResKey.ptrValue)]
				if !bHas {
					return nil, fmt.Errorf("no item named %s in the json", (*(*string)(pResKey.ptrValue)))
				}
			} else if '0' <= item && '9' >= item {
				if vArrayType != res.vType {
					return nil, errors.New("path is not match with type, expect array")
				}
				reader.UnreadByte()
				pathIndex := getPathIndex(reader)
				if pathIndex >= len(((*tJsonArray)(res.ptrValue)).lstValues) {
					return nil, fmt.Errorf("array index %d is out of range", pathIndex)
				}
				res = ((*tJsonArray)(res.ptrValue)).lstValues[pathIndex]
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
				return nil, errors.New("path ends with '.' or too many '.'")
			}
		} else {
			return nil, fmt.Errorf("invalid char %c in the path", item)
		}
	}
	return res, nil
}

//get the original string of the specified path
func (gson *Gson) Original(strPath string) (string, error) {
	value, err := gson.Get(strPath)
	if nil != err {
		return "", err
	}
	return value.Dump(), nil
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
		if err := recover(); nil != err {
			resErr = fmt.Errorf("%v", err)
		}
	}()
	value, err := gson.Get(strPath)
	if nil != err {
		return err
	}

	newVal := getValueByString(strOrigVal)
	if nil != newVal {
		*value = *newVal
		return nil
	}
	return errors.New("set failed")
}

//AddObject should have path, the item related with the path should be an object
//There should also be key and value in the function
func (gson *Gson) AddObject(strPath, strKey, strOrigVal string) (resErr error) {
	defer func() {
		if err := recover(); nil != err {
			resErr = fmt.Errorf("%v", err)
		}
	}()
	value, err := gson.Get(strPath)
	if nil != err {
		return err
	}
	if vObjectType != value.vType {
		return errors.New("the item with specified path should be an object")
	}
	_, ok := ((*tJsonObject)(value.ptrValue)).mapObjects[strKey]
	if ok {
		return errors.New("the key already exist")
	}
	newVal := getValueByString(strOrigVal)
	if nil == newVal {
		return errors.New("invalid value format")
	}
	((*tJsonObject)(value.ptrValue)).mapObjects[strKey] = newVal
	return nil
}

//AddMember should have path, the item related with the path should be an array
//There should also be a value in the function
func (gson *Gson) AddMember(strPath, strOrigVal string) (resErr error) {
	defer func() {
		if err := recover(); nil != err {
			resErr = fmt.Errorf("%v", err)
		}
	}()
	value, err := gson.Get(strPath)
	if nil != err {
		return err
	}
	if vArrayType != value.vType {
		return errors.New("the item with specified path should be an array")
	}
	newVal := getValueByString(strOrigVal)
	if nil == newVal {
		return errors.New("invalid value format")
	}
	((*tJsonArray)(value.ptrValue)).lstValues = append(((*tJsonArray)(value.ptrValue)).lstValues, newVal)
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
			panic("empty index")
		}
		if '0' <= item && '9' >= item {
			resIndex = 10*resIndex + (int)(item-'0')
			bHasValue = true
		} else if PATH_SPLIT_CHAR == item {
			reader.UnreadByte()
			break
		} else {
			panic("unexpected char in the array index")
		}
	}
	return resIndex
}

func parse(reader *bufio.Reader) *Value {
	escapeWhiteSpace(reader)
	//the format of json is invalid
	item, err := reader.ReadByte()
	if nil != err {
		panic("empty json string")
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
	res := &Value{vObjectType, unsafe.Pointer(new(tJsonObject))}
	((*tJsonObject)(res.ptrValue)).mapObjects = map[string]*Value{}
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
		_, ok := ((*tJsonObject)(res.ptrValue)).mapObjects[*pStrKey]
		if ok {
			panic("duplicate key in json object")
		}
		((*tJsonObject)(res.ptrValue)).mapObjects[*pStrKey] = parseValue(reader)
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
	res := &Value{vArrayType, unsafe.Pointer(new(tJsonArray))}
	//empty array
	if ']' == item {
		return res
	}
	reader.UnreadByte()
	for {
		//will take the next item
		val := parseValue(reader)
		((*tJsonArray)(res.ptrValue)).lstValues = append(((*tJsonArray)(res.ptrValue)).lstValues, val)
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
			panic("string not finished")
		}
		if '"' == item {
			strValue = string(byteRes)
			res.ptrValue = unsafe.Pointer(&strValue)
			break
		}
		if '\\' == item {
			nextItem, err = reader.ReadByte()
			if nil != err {
				panic("invalid end character: \\")
			}
			value, ok := byteSpecial[nextItem]
			if ok {
				byteRes = append(byteRes, value)
				continue
			} else {
				panic("invalid escape character")
			}
		}
		byteRes = append(byteRes, item)
	}
	return res
}

func parseNil(reader *bufio.Reader) *Value {
	items, err := reader.Peek(3)
	if nil != err || !('u' == items[0] && 'l' == items[1] && 'l' == items[2]) {
		panic("string should be started with quotation marks: " + string(items))
	}
	res := &Value{}
	res.vType = vNilType
	reader.Discard(3)
	return res
}

func parseTrue(reader *bufio.Reader) *Value {
	items, err := reader.Peek(3)
	if nil != err || !('r' == items[0] && 'u' == items[1] && 'e' == items[2]) {
		panic("string should be started with quotation marks: " + string(items))
	}
	res := &Value{}
	res.vType = vTrueType
	reader.Discard(3)
	return res
}

func parseFalse(reader *bufio.Reader) *Value {
	items, err := reader.Peek(4)
	if nil != err || !('a' == items[0] && 'l' == items[1] && 's' == items[2] && 'e' == items[3]) {
		panic("string should be started with quotation marks: " + string(items))
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
		return &Value{vNumType, unsafe.Pointer(&tNumber{vIntType, unsafe.Pointer(&resNum), &strNum})}
	}
	//maybe number exceed int range
	if resNum, err := strconv.ParseUint(strNum, 10, 64); nil == err {
		return &Value{vNumType, unsafe.Pointer(&tNumber{vUIntType, unsafe.Pointer(&resNum), &strNum})}
	}
	if resNum, err := strconv.ParseFloat(strNum, 64); nil == err {
		return &Value{vNumType, unsafe.Pointer(&tNumber{vDoubleType, unsafe.Pointer(&resNum), &strNum})}
	}
	panic("invalid number")
}
