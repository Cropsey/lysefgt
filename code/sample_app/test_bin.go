// Sample application whose only purpose is to be in go and trivial to disassemble
// when figuring out how to instrument with ebpf

package main

import (
	"time"
)

func easyToFindFunctionName() {
	time.Sleep(5 * time.Second)
}

func alsoEasyToFindFunctionName() {
	time.Sleep(5 * time.Second)
}

func main() {
	for {
		easyToFindFunctionName()
		alsoEasyToFindFunctionName()
	}
}
