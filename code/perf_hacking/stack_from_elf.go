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
	file   string
	line   int
	symbol string
}

func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x 0x%0.8x %v:%v %v()", sp.addr, sp.pc, sp.file, sp.line, sp.symbol)
}

type elfHelper struct {
	symbols []elf.Symbol
	elfFile *elf.File
}

func newElf(binPath string) elfHelper {
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

func (e elfHelper) seekDwarfEntry(symbol string) (*stackPos, error) {
	// TODO: figure out how to cache these
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

func (e elfHelper) humanReadableStack(stack [100]uint64) []stackPos {
	prettyStack := []stackPos{}
	for _, addr := range stack {
		if addr == 0 {
			break
		}
		prev := stackPos{addr: addr}
		for _, s := range e.symbols {
			if addr < s.Value {
				entry, err := e.seekDwarfEntry(prev.symbol)
				if err != nil {
					log.Printf("failed to dwarf SeekToTypeNamed(%v) = %v\n", prev.symbol, err)
				} else {
					if entry != nil {
						prev.file = entry.file
						prev.line = entry.line
					}
				}
				prettyStack = append(prettyStack, prev)
				break
			}
			prev.symbol = s.Name
			prev.pc = s.Value
		}
	}
	return prettyStack
}
