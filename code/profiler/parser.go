package main

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
	"sort"
)

// To keep track of particular stack trace position enhanced with human readable metadata
type stackPos struct {
	addr   uint64 // Exact address from the stack at the time of perf event sample
	pc     uint64 // Address of the symbol for the exact address
	symbol string // Symbol from ELF
}

// Prettier print for stackPos
func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x 0x%0.8x %-33v", sp.addr, sp.pc, sp.symbol+"()")
}

// Pre-processed ELF data of the profiled binary
type elfHelper struct {
	symbols []elf.Symbol
	elfFile *elf.File
}

// Initialize ELF helper
func newElf(pid int) elfHelper {
	// Figure out the binary path for the process with PID
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

	// Get symbols from ELF and sort them by address
	e.symbols, err = ef.Symbols()
	if err != nil {
		log.Fatalf("failed to get elf symbols for file %q: %v\n", binPath, err)
	}
	sort.Slice(e.symbols, func(i, j int) bool { return e.symbols[i].Value < e.symbols[j].Value })
	e.elfFile = ef
	return e
}

// Convert the address stack trace to human readable stack trace
func (e elfHelper) humanReadableStack(stack [10]uint64) []stackPos {
	st := []stackPos{}
	// For each address in the stack trace
	for _, addr := range stack {
		if addr == 0 {
			break
		}
		prev := stackPos{addr: addr}
		// Find the matching symbol from ELF
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

// Convert 0 terminated C string to Go string
func (event bpfEvent) taskComm() string {
	bs := make([]byte, 0, len(event.Name))
	for i := 0; i < len(event.Name) && event.Name[i] != 0; i++ {
		bs = append(bs, byte(event.Name[i]))
	}
	return string(bs)
}
