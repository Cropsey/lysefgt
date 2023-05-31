package main

import "fmt"

// To keep track of particular stack trace position
type stackPos struct {
	addr uint64 // Exact address from the stack at the time of perf event sample
}

// Prettier print for stackPos
func (sp stackPos) String() string {
	return fmt.Sprintf("0x%0.8x", sp.addr)
}

// Convert the address stack trace to human readable stack trace
func stack(stack [10]uint64) []stackPos {
	st := []stackPos{}
	// For each address in the stack trace
	for _, addr := range stack {
		if addr == 0 {
			break
		}
		st = append(st, stackPos{addr: addr})
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
