// Sample application whose only purpose is to be in go and trivial to disassemble
// when figuring out how to instrument with ebpf

package main

import "fmt"

const x = 100000

func easyToFindFunctionName() {
	fmt.Println("easyToFindFunctionName")
	sum := 0
	for i := 0; i < x; i++ {
		for j := 0; j < x; j++ {
			sum += i + j
		}
	}
}

func alsoEasyToFindFunctionName() {
	fmt.Println("alsoEasyToFindFunctionName")
	sum := 0
	for i := 0; i < x; i++ {
		for j := 0; j < x/10; j++ {
			sum += i + j
		}
	}
}

func main() {
	for {
		easyToFindFunctionName()
		alsoEasyToFindFunctionName()
	}
}
