package main

import "fmt"

type stackPos struct {
	addr uint64
}

func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x", sp.addr)
}

func stack(stack [100]uint64) []stackPos {
	st := []stackPos{}
	for _, addr := range stack {
		if addr == 0 {
			break
		}
		st = append(st, stackPos{addr: addr})
	}
	return st
}
