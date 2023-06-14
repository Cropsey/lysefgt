package main

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/go-delve/delve/pkg/dwarf/reader"
)

// To keep track of particular stack trace position enhanced with human readable metadata
type stackPos struct {
	addr   uint64 // Exact address from the stack at the time of perf event sample
	pc     uint64 // Address of the symbol for the exact address
	symbol string // Symbol from ELF
	file   string // Source code file path from DWARF
	line   int    // Source code file line from DWARF
}

// Prettier print for stackPos
func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x 0x%0.8x %-33v  %v:%v", sp.addr, sp.pc, sp.symbol+"()", sp.file, sp.line)
}

// Pre-processed ELF data of the profiled binary
type elfHelper struct {
	symbols []elf.Symbol
	elfFile *elf.File
}

// Statistics for stack trace symbol counts
type stats struct {
	count map[string]int
	meta  map[string]stackPos
}

// Helper for sorting symbols from stack trace stats by most frequently occurring
type sortSymbolCount struct {
	symbol string
	count  int
}

// Initialize stats counters for stack trace symbol counts
func newStats() *stats {
	return &stats{
		count: make(map[string]int),
		meta:  make(map[string]stackPos),
	}
}

// Process stack position by stats counter
func (a *stats) add(sp stackPos) {
	a.count[sp.symbol] += 1
	if _, exists := a.meta[sp.symbol]; !exists {
		a.meta[sp.symbol] = sp
	}
}

// Print stack trace stats summary
func (a *stats) summary() {
	var s []sortSymbolCount
	for symbol, count := range a.count {
		s = append(s, sortSymbolCount{symbol: symbol, count: count})
	}
	sort.Slice(s, func(i, j int) bool { return s[i].count > s[j].count })
	fmt.Println()
	fmt.Println("AGGREGATED PERF EVENT SAMPLES:")
	fmt.Println("  COUNT  SYMBOL                                 FILE:LINE")
	fmt.Println("  -----  -------------------------------------  ------------------------------------")
	for _, sorted := range s {
		sp := a.meta[sorted.symbol]
		fmt.Printf("%7d  %-37v  %v:%v\n", sorted.count, sp.symbol+"()", sp.file, sp.line)
	}
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
				// Enhance by DWARF data if available
				entry, err := e.seekDwarfEntry(prev.symbol)
				if err == nil && entry != nil {
					prev.file = entry.file
					prev.line = entry.line
				}
				st = append(st, prev)
				break
			}
			prev.symbol = s.Name
			prev.pc = s.Value
		}
	}
	return st
}

func (e elfHelper) seekDwarfEntry(symbol string) (*stackPos, error) {
	// Calling this for every event is horribly inefficient, ideally this should be lazily cached in a lookup table
	dwrf, err := e.elfFile.DWARF()
	if err != nil {
		return nil, fmt.Errorf("failed to get dwarf: %w", err)
	}
	dwr := reader.New(dwrf)
	// Keep track of last compile unit so when subprogram is found, line entry reader can be initialized
	var lastCompileUnit *dwarf.Entry
	// Iterate over DWARF entries
	for entry, err := dwr.Next(); entry != nil; entry, err = dwr.Next() {
		if err != nil {
			return nil, fmt.Errorf("failed to get next DWARF entry: %w", err)
		}
		// The entry is a CompileUnit
		if entry.Tag == dwarf.TagCompileUnit {
			pin := *entry
			lastCompileUnit = &pin
		}
		// It has a DWARF attribute "name"
		name, ok := entry.Val(dwarf.AttrName).(string)
		if !ok {
			continue
		}
		// The DWARF attribute "name" matches ELF symbol
		if name == symbol {
			// Create DWARF reader for the line table
			lr, err := dwrf.LineReader(lastCompileUnit)
			if err != nil {
				return nil, fmt.Errorf("failed to create DWARF line reader: %w", err)
			}
			le := dwarf.LineEntry{}
			if entry.Val(dwarf.AttrLowpc) == nil {
				return nil, fmt.Errorf("DWARF entry has no LowPC attribute")
			}
			pc := entry.Val(dwarf.AttrLowpc).(uint64)
			// Find the line entry
			if err := lr.SeekPC(pc, &le); err != nil {
				return nil, fmt.Errorf("failed to seek DWARF line entry at pc %v: %w", pc, err)
			}
			// Add source code file name and line to the stack trace
			return &stackPos{line: le.Line, file: le.File.Name}, nil
		}
	}

	return nil, fmt.Errorf("entry %v not found", symbol)
}

// Convert 0 terminated C string to Go string
func (event bpfEvent) taskComm() string {
	bs := make([]byte, 0, len(event.Name))
	for i := 0; i < len(event.Name) && event.Name[i] != 0; i++ {
		bs = append(bs, byte(event.Name[i]))
	}
	return string(bs)
}
