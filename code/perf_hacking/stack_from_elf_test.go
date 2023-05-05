package main

import (
	"log"
	"testing"
)

func TestStackPretty(t *testing.T) {
	stack := [100]uint64{
		0x00439687, 0x0043ac3e, 0x0043b16d, 0x0045d5e3, 0x0045c735, 0x0048ea47, 0x0048eb19, 0x0048ebe5, 0x00434592, 0x0045f6e1,
		0x0048ea47,
	}

	elf := newElf("../report_function_call/testbin")
	ustack := elf.humanReadableStack(stack)
	for _, l := range ustack {
		log.Printf("  %v\n", l)
	}
}
