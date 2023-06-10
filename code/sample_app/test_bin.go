// Sample application whose only purpose is to be in go and trivial to disassemble
// when figuring out how to instrument with ebpf

package main

import "fmt"

const x = 100000

func easyToFindFunctionName20() { easyToFindFunctionName19() }
func easyToFindFunctionName19() { easyToFindFunctionName18() }
func easyToFindFunctionName18() { easyToFindFunctionName17() }
func easyToFindFunctionName17() { easyToFindFunctionName16() }
func easyToFindFunctionName16() { easyToFindFunctionName15() }
func easyToFindFunctionName15() { easyToFindFunctionName14() }
func easyToFindFunctionName14() { easyToFindFunctionName13() }
func easyToFindFunctionName13() { easyToFindFunctionName12() }
func easyToFindFunctionName12() { easyToFindFunctionName11() }
func easyToFindFunctionName11() { go easyToFindFunctionName10() }
func easyToFindFunctionName10() { easyToFindFunctionName9() }
func easyToFindFunctionName9()  { easyToFindFunctionName8() }
func easyToFindFunctionName8()  { easyToFindFunctionName7() }
func easyToFindFunctionName7()  { easyToFindFunctionName6() }
func easyToFindFunctionName6()  { easyToFindFunctionName5() }
func easyToFindFunctionName5()  { easyToFindFunctionName4() }
func easyToFindFunctionName4()  { easyToFindFunctionName3() }
func easyToFindFunctionName3()  { easyToFindFunctionName2() }
func easyToFindFunctionName2()  { easyToFindFunctionName1() }

func easyToFindFunctionName1() {
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
		for j := 0; j < x; j++ {
			sum += i + j
		}
	}
}

func main() {
	for {
		easyToFindFunctionName20()
		//alsoEasyToFindFunctionName()
	}
}
