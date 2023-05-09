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

type stackPos struct {
	addr   uint64
	pc     uint64
	symbol string
	file   string
	line   int
}

func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x 0x%0.8x %-33v  %v:%v", sp.addr, sp.pc, sp.symbol+"()", sp.file, sp.line)
}

type elfHelper struct {
	symbols []elf.Symbol
	elfFile *elf.File
}

type aggregate struct {
	count map[string]int
	meta  map[string]stackPos
}

type sortSymbolCount struct {
	symbol string
	count  int
}

func newAggregate() *aggregate {
	return &aggregate{
		count: make(map[string]int),
		meta:  make(map[string]stackPos),
	}
}

func (a *aggregate) add(sp stackPos) {
	a.count[sp.symbol] += 1
	if _, exists := a.meta[sp.symbol]; !exists {
		a.meta[sp.symbol] = sp
	}
}

func (a *aggregate) print() {
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
				entry, err := e.seekDwarfEntry(prev.symbol)
				if err == nil && entry != nil {
					prev.file = entry.file
					prev.line = entry.line
					st = append(st, prev)
				}
				break
			}
			prev.symbol = s.Name
			prev.pc = s.Value
		}
	}
	return st
}

func (e elfHelper) seekDwarfEntry(symbol string) (*stackPos, error) {
	// calling this for every event is horribly inefficient, ideally this should be cached in a lookup table
	dwrf, err := e.elfFile.DWARF()
	if err != nil {
		return nil, fmt.Errorf("failed to get dwarf: %w", err)
	}
	dwr := reader.New(dwrf)
	// keep track of last compile unit so when subprogram is found, line entry reader can be initialized
	var lastCompileUnit *dwarf.Entry
	for entry, err := dwr.Next(); entry != nil; entry, err = dwr.Next() {
		if err != nil {
			return nil, fmt.Errorf("failed to get next DWARF entry: %w", err)
		}
		if entry.Tag == dwarf.TagCompileUnit {
			pin := *entry
			lastCompileUnit = &pin
		}
		name, ok := entry.Val(dwarf.AttrName).(string)
		if !ok {
			continue
		}
		if name == symbol {
			lr, err := dwrf.LineReader(lastCompileUnit)
			if err != nil {
				return nil, fmt.Errorf("failed to create DWARF line reader: %w", err)
			}
			le := dwarf.LineEntry{}
			pc := entry.Val(dwarf.AttrLowpc).(uint64)
			if err := lr.SeekPC(pc, &le); err != nil {
				return nil, fmt.Errorf("failed to seek DWARF line entry at pc %v: %w", pc, err)
			}
			return &stackPos{line: le.Line, file: le.File.Name}, nil
		}
	}

	return nil, fmt.Errorf("entry %v not found", symbol)
}

// convert 0 terminated string to go string
func (event bpfEvent) taskComm() string {
	bs := make([]byte, 0, len(event.Name))
	for i := 0; i < len(event.Name) && event.Name[i] != 0; i++ {
		bs = append(bs, byte(event.Name[i]))
	}
	return string(bs)
}
