// Sample application whose only purpose is to be in go and trivial to disassemble
// when figuring out how to instrument with ebpf

package main

import (
	"fmt"
	"time"
)

func easyToFindFunctionName(arg uint32) {
	time.Sleep(5 * time.Second)
	fmt.Println(arg)
}

func easyToFindFunctionNameNoArg() {
	fmt.Println("")
}

func EasyToFindFunctionName(arg uint32) {
	easyToFindFunctionName(arg)
}

func main() {
	easyToFindFunctionNameNoArg()
	t1 := time.NewTicker(time.Second * 3)
	t2 := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t1.C:
			EasyToFindFunctionName(1)
		case <-t2.C:
			EasyToFindFunctionName(2)
		}
	}
}
