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

func isWhiteSpace(strItem string) bool {
	return " " == strItem || "\t" == strItem || "\n" == strItem
}

func escapeWhiteSpace(scanner *bufio.Scanner) {
	for scanner.Scan() {
		if isWhiteSpace(scanner.Text()) {
			continue
		}
	}
}

func parseValue(scanner *bufio.Scanner) Value {
	switch scanner.Text() {
	case "{":
		parseObject(scanner)
	case "[":
		parseArray(scanner)
	case `"`:
		parseString(scanner)
	case "n":
		parseNil(scanner)
	case "t":
		parseTrue(scanner)
	case "f":
		parseFalse(scanner)
	default:
		parseNumber(scanner)
	}
}

func parse(scanner *bufio.Scanner) *Value {
	scanner.Split(bufio.ScanBytes)
	escapeWhiteSpace(scanner)
	//the format of json is invalid
	if "{" != scanner.Text() {
		panic("json string not started with {")
	}
	return parseObject(scanner)
}

func parseObject(scanner *bufio.Scanner) *Value {
	res := &Value{vObjectType, unsafe.Pointer(new(JsonObject))}
	//empty objext
	if "}" == scanner.Text() {
		return res
	}
	for {
		//will take the next item
		escapeWhiteSpace(scanner)
		if `"` != scanner.Text() {
			panic("json key is not a string.")
		}
		key := parseString(scanner)
		pStrKey := (*string)(key.ptrValue)
		escapeWhiteSpace(scanner)
		if ":" != scanner.Text() {
			panic("invalid json format")
		}
		escapeWhiteSpace(scanner)
		val := parseValue(scanner)
		((*JsonObject)(res.ptrValue)).lstObjects = append(((*JsonObject)(res.ptrValue)).lstObjects, Member{pStrKey, val})
		//another member, do not take next item first
		escapeWhiteSpace(scanner)
		if "," == scanner.Text() {
			continue
		}
		if "}" == scanner.Text() {
			break
		}
		panic("invalid json format")
	}
	return res
}

func parseArray(scanner *bufio.Scanner) *Value {
	res := &Value{vArrayType, unsafe.Pointer(new(JsonArray))}
	//empty array
	if "]" == scanner.Text() {
		return res
	}
	for {
		//will take the next item
		escapeWhiteSpace(scanner)
		val := parseValue(scanner)
		((*JsonArray)(res.ptrValue)).lstValues = append(((*JsonArray)(res.ptrValue)).lstValues, val)
		escapeWhiteSpace(scanner)
		//another member, do not take next item first
		if "," == scanner.Text() {
			continue
		}
		if "]" == scanner.Text() {
			break
		}
		panic("invalid json format")
	}
	return res
}

func parseString(scanner *bufio.Scanner) *Value {

}

func parseNil(scanner *bufio.Scanner) *Value {
	res := &Value{}
	res.vType = vNilType
	if !(scanner.Scan() && "i" == scanner.Text() &&
		scanner.Scan() && "l" == scanner.Text()) {
		panic("invalid value")
	}
	return res
}

func parseTrue(scanner *bufio.Scanner) *Value {
	res := &Value{}
	res.vType = vTrueType
	if !(scanner.Scan() && "r" == scanner.Text() &&
		scanner.Scan() && "u" == scanner.Text() &&
		scanner.Scan() && "e" == scanner.Text()) {
		panic("invalid value")
	}
	return res
}

func parseFalse(scanner *bufio.Scanner) *Value {
	res := &Value{}
	res.vType = vFalseType
	if !(scanner.Scan() && "a" == scanner.Text() &&
		scanner.Scan() && "l" == scanner.Text() &&
		scanner.Scan() && "s" == scanner.Text() &&
		scanner.Scan() && "e" == scanner.Text()) {
		panic("invalid value")
	}
	return res
}

func parseNumber(scanner *bufio.Scanner) *Value {

}
