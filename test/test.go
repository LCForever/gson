package main

import (
	"bufio"
	"strings"

	"../../gson/gson"
)

func main() {
	reader := bufio.NewReader(strings.NewReader(`{"json": [123, 345, [1,2,3],	{"abc":false}]}}`))
	gson.Parse(reader)
}
