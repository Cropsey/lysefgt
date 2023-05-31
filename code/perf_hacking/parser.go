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

// convert 0 terminated string to go string
func (event bpfEvent) taskComm() string {
	bs := make([]byte, 0, len(event.Name))
	for i := 0; i < len(event.Name) && event.Name[i] != 0; i++ {
		bs = append(bs, byte(event.Name[i]))
	}
	return string(bs)
}
