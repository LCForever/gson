# gson
gson is a Go package which provides a convenient way to do operations in a json document. It has features such as parse/modify json documents, dot natation paths, etc.
## Parse json string
The Parse function parses the json string in a dom style. The result is saved in a object named Gson.
```go
package main
import (
	"bufio"
	"fmt"
	"strings"
	"github.com/LCForever/gson"
)
func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 123, 345, [1,2,3],	{"abc\n\"":false}]}`))
	MyGson := new(gson.Gson)
	_, err := MyGson.Parse(reader)
	if nil != err{
		fmt.Println(err)
	}else{
		fmt.Println("json parse succeed")
	}
}
```
This will print:
```
json parse succeed
```
If the format of the json string is not correct, parse will fail and an error will return
```go
package main
import (
	"bufio"
	"fmt"
	"strings"
	"github.com/LCForever/gson"
)
func main() {
	//the last } is deleted
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 123, 345, [1,2,3],	{"abc\n\"":false}]`))
	MyGson := new(gson.Gson)
	_, err := MyGson.Parse(reader)
	if nil != err{
		fmt.Println(err)
	}else{
		fmt.Println("json parse succeed")
	}
}
```
This will print:
```
json string not finished
```
## Get value by path
This pacakge supplies a way to obtain the value by dot natation path.
```go
package main
import (
	"bufio"
	"fmt"
	"strings"
	"github.com/LCForever/gson"
)
func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 123, 345, [1,2,3],	{"abc\n\"":false}]}`))
	MyGson := new(gson.Gson)
	_, err := MyGson.Parse(reader)
	if nil != err{
		fmt.Println(err)
		return
	}
	//0 means index of an array
	val, err := MyGson.Get(`"json".0`)
	val2, err2 := MyGson.Get(`"json"."abc\n\"`)
	if nil != err || nil == val || nil != err2 || nil == val2 {
		fmt.Println("get value failed")
		return
	}
	fmt.Println(val.Dump())
	fmt.Println(val2.Dump())
}
```
This will print
```
"111"
false
```
## update value by path
```go
package main
import (
	"bufio"
	"fmt"
	"strings"
	"github.com/LCForever/gson"
)
func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 123, 345, [1,2,3],	{"abc\n\"":false}]}`))
	MyGson := new(gson.Gson)
	_, err := MyGson.Parse(reader)
	if nil != err{
		fmt.Println(err)
		return
	}
	MyGson.Set(`"json".0`, `234`)
	MyGson.Set(`"json".1`, `"hello world"`)
	MyGson.Dump()
}
```
This will print
```
{"json": [234,"hello world",345,[1,2,3],{"abc\n\"":false}]}
```
## add value by path
```go
package main
import (
	"bufio"
	"fmt"
	"strings"
	"github.com/LCForever/gson"
)
func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 123, 345, [1,2,3],	{"abc\n\"":false}]}`))
	MyGson := new(gson.Gson)
	_, err := MyGson.Parse(reader)
	if nil != err{
		fmt.Println(err)
		return
	}
	MyGson.AddMember(`"json"`, `{"test": "value"}`)
	MyGson.AddObject(`"json".5`, `"key"`, `"hello world"`)
	MyGson.Dump()
}
```
This will print
```
{"json": ["111",123,345,[1,2,3],{"abc\n\"":false},{"test":"value", "key":"hello world"}]}
```
## Some other useful functions
```go
//Get the value of an item
value.GetArrayValue() >> []*Value, error
value.GetObjectValue() >> map[string]*Value, error
value.GetBoolValue() >> bool, error
value.GetIntValue() >> int64, error
value.GetUIntValue() >> uint64, error
value.GetDoubleValue() >> float64, error
value.GetStringValue() >> string, error
//Check the value of an item
value.IsObject() >> bool
value.IsArray() >> bool
value.IsNil() >> bool
value.IsNumber() >> bool
value.IsString() >> bool
value.IsBool() >> bool
//Dump the value to a string
value.Dump() >> string
```
