package main

import (
	"bufio"
	"fmt"
	"strings"

	"../../gson/gson"
)

func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": ["111", 36000000000000000000000000.4, 123, 345, [1,2,3],	{"abc\n\"":false}]}`))
	MyGson := new(gson.Gson)
	MyGson.Parse(reader)
	fmt.Println(MyGson.Dump())
	fmt.Println(MyGson.AddMember(`"json"`, "123456"))
	fmt.Println(MyGson.Dump())
	fmt.Println(MyGson.Set(`"json"`, "123456"))
	fmt.Println(MyGson.Dump())
	fmt.Println(MyGson.AddObject(`"json"`, "newitem", `"123456"`))
	fmt.Println(MyGson.Dump())
}
