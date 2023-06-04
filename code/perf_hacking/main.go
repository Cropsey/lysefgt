package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target native -type event bpf perf.c -- -I../headers

func recordBuffer(rd *perf.Reader, recordsChan chan *perf.Record) {
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				recordsChan <- nil
			}
			continue
		}
		recordsChan <- &record
	}
}

func main() {
	var pid int
	var verbose int
	flag.IntVar(&pid, "pid", 0, "PID of the profiled process")
	flag.IntVar(&verbose, "v", 0, "Verbosity of perf event logs, 0 prints only aggregate, 1 prints each perf event record")
	flag.Parse()
	log.SetOutput(os.Stderr)

	// Leverage cilium/ebpf generated scaffold
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %s", err)
	}
	defer objs.Close()

	// Open a perf reader from userspace into the perf event array created earlier.
	rd, err := perf.NewReader(objs.Events, os.Getpagesize())
	if err != nil {
		log.Fatalf("Creating event reader: %s", err)
	}
	defer rd.Close()
	recordsChan := make(chan *perf.Record, 1024)
	go recordBuffer(rd, recordsChan)

	log.Println("Waiting for events..")

	anyCPU := -1
	groupLeader := -1
	// perf_event_open syscall
	perfEventFD, err := unix.PerfEventOpen(
		&unix.PerfEventAttr{
			Type:        unix.PERF_TYPE_SOFTWARE,
			Config:      unix.PERF_COUNT_SW_CPU_CLOCK,
			Sample_type: unix.PERF_SAMPLE_RAW,
			Sample:      1,
			Wakeup:      1,
		},
		pid,
		anyCPU,
		groupLeader,
		unix.PERF_FLAG_FD_CLOEXEC,
	)
	if err != nil {
		log.Fatalf("Failed to create the perf event: %v", err)
	}
	defer func() {
		if err := unix.Close(perfEventFD); err != nil {
			log.Fatalf("Failed to close perf event: %v", err)
		}
	}()

	// attach ebpf to perf event
	if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_SET_BPF, objs.bpfPrograms.PerfEventStacktrace.FD()); err != nil {
		log.Fatalf("Failed to attach eBPF program to perf event: %v", err)
	}

	// start perf event
	if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_ENABLE, 0); err != nil {
		log.Fatalf("Failed to enable perf event: %v", err)
	}

	defer func() {
		if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_DISABLE, 0); err != nil {
			log.Fatalf("Failed to disable perf event: %v", err)
		}
	}()

	elf := newElf(pid)
	var event bpfEvent
	var record perf.Record
	aggregate := newAggregate()
	defer aggregate.print()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			return
		case r := <-recordsChan:
			if r == nil {
				return
			}
			record = *r
		}
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parsing perf event failed: %s", err)
		}
		ustack := elf.humanReadableStack(event.UserStack)
		for _, l := range ustack {
			aggregate.add(l)
		}
		if verbose >= 1 {
			fmt.Printf("bin[%v] pid[%v]\n", event.taskComm(), pid)
			fmt.Println("  ADDRESS    PC         SYMBOL                             FILE:LINE")
			fmt.Println("  ---------  ---------  ---------------------------------  ------------------------------------")
			for _, l := range ustack {
				fmt.Println(" ", l)
			}
			fmt.Println()
		}
	}
}
