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
	"syscall"

	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

// generate
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target native -type event bpf perf.c -- -I../headers

func main() {
	var pid int
	flag.IntVar(&pid, "pid", 0, "PID of the profiled process")
	flag.Parse()
	log.SetOutput(os.Stderr)

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %s", err)
	}
	defer objs.Close()
	// Subscribe to signals for terminating the program.
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Open a perf reader from userspace into the perf event array
	// created earlier.
	rd, err := perf.NewReader(objs.Events, os.Getpagesize())
	if err != nil {
		log.Fatalf("Creating event reader: %s", err)
	}
	defer rd.Close()

	// Close the reader when the process receives a signal, which will exit
	// the read loop.
	go func() {
		<-stopper
		rd.Close()
	}()

	anyCPU := -1
	groupLeader := -1
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

	binPath, err := os.Readlink(fmt.Sprintf("/proc/%v/exe", pid))
	if err != nil {
		log.Fatalf("Failed to read binpath from pid %v: %v", err)
	}
	elf := newElf(binPath)
	var event bpfEvent
	log.Println("Waiting for events..")

	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, perf.ErrClosed) {
				log.Println("Received signal, exiting..")
				return
			}
			log.Printf("Reading from reader: %s", err)
			continue
		}
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parsing perf event error: %s", err)
		}

		fmt.Printf("bin[%v] pid[%v]\n", event.taskComm(), event.Pid)
		ustack := elf.humanReadableStack(event.UserStack)
		fmt.Println("  ADDRESS    PC         SYMBOL                             FILE:LINE")
		fmt.Println("  ---------  ---------  ---------------------------------  ------------------------------------")
		for _, l := range ustack {
			fmt.Println(" ", l)
		}
		fmt.Println()
	}
}
