package main

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
	"sort"
)

type stackPos struct {
	addr   uint64
	pc     uint64
	symbol string
}

func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x 0x%0.8x %-33v", sp.addr, sp.pc, sp.symbol+"()")
}

type elfHelper struct {
	symbols []elf.Symbol
	elfFile *elf.File
}

func newElf(pid int) elfHelper {
	binPath, err := os.Readlink(fmt.Sprintf("/proc/%v/exe", pid))
	if err != nil {
		log.Fatalf("Failed to read binpath from pid %v: %v", err)
	}
	f, err := os.Open(binPath)
	if err != nil {
		log.Fatalf("failed to open file %q: %v\n", binPath, err)
	}
	ef, err := elf.NewFile(f)
	if err != nil {
		log.Fatalf("failed to elf open file %q: %v\n", binPath, err)
	}
	e := elfHelper{}
	e.symbols, err = ef.Symbols()
	if err != nil {
		log.Fatalf("failed to get elf symbols for file %q: %v\n", binPath, err)
	}
	sort.Slice(e.symbols, func(i, j int) bool { return e.symbols[i].Value < e.symbols[j].Value })
	e.elfFile = ef
	return e
}

func (e elfHelper) humanReadableStack(stack [100]uint64) []stackPos {
	st := []stackPos{}
	for _, addr := range stack {
		if addr == 0 {
			break
		}
		prev := stackPos{addr: addr}
		for _, s := range e.symbols {
			if addr < s.Value {
				st = append(st, prev)
				break
			}
			prev.symbol = s.Name
			prev.pc = s.Value
		}
	}
	return st
}

// convert 0 terminated string to go string
func (event bpfEvent) taskComm() string {
	bs := make([]byte, 0, len(event.Name))
	for i := 0; i < len(event.Name) && event.Name[i] != 0; i++ {
		bs = append(bs, byte(event.Name[i]))
	}
	return string(bs)
}
