package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

// Use cilium/ebpf package to generate eBPF scaffold for programs and maps
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target native -type event bpf perf.c -- -I../headers

// main is an entrypoint to the profiler
func main() {
	// Command line flag parsing
	var pid int
	flag.IntVar(&pid, "pid", 0, "PID of the profiled process")
	flag.Parse()

	// Print errors where they belong
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

	log.Println("Waiting for events..")

	anyCPU := -1      // Sample application on any CPU
	groupLeader := -1 // Disable event grouping
	// Go wrapper for perf_event_open syscall
	perfEventFD, err := unix.PerfEventOpen(
		&unix.PerfEventAttr{
			Type:        unix.PERF_TYPE_SOFTWARE,        // Indicates software-defined event, defines available values for `Config`
			Config:      unix.PERF_COUNT_SW_CPU_CLOCK,   // Reports the CPU clock, high-resolution per-CPU timer, connected to `Type` defined above
			Sample_type: unix.PERF_SAMPLE_RAW,           // Allowing eBPF to record additional data
			Sample:      uint64(time.Millisecond * 100), // Create perf event sample every 100ms
			Wakeup:      1,                              // Don't skip any samples
		},
		pid,                       // Profiled application PID, must be present given anyCPU argument value -1
		anyCPU,                    // The CPU ID, -1 means all CPUs
		groupLeader,               // Configure event grouping, -1 means disabled
		unix.PERF_FLAG_FD_CLOEXEC, // Close the event fd on execve, avoiding possible race conditions
	)
	if err != nil {
		log.Fatalf("Failed to create the perf event: %v", err)
	}
	defer func() {
		if err := unix.Close(perfEventFD); err != nil {
			log.Fatalf("Failed to close perf event: %v", err)
		}
	}()

	// Tell kernel to attach eBPF program to the perf event fd
	if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_SET_BPF, objs.bpfPrograms.PerfEventStacktrace.FD()); err != nil {
		log.Fatalf("Failed to attach eBPF program to perf event: %v", err)
	}

	// Tell kernel to enable the perf event
	if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_ENABLE, 0); err != nil {
		log.Fatalf("Failed to enable perf event: %v", err)
	}

	defer func() {
		if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_DISABLE, 0); err != nil {
			log.Fatalf("Failed to disable perf event: %v", err)
		}
	}()

	// Forward declarations for the loop variables
	var event bpfEvent

	// Loop forever and process stack traces
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

		// Read eBPF binary data into go native structure, the ABI compatibility is ensured by cilium/ebpf
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parsing perf event failed: %s", err)
		}

		// Get the stack trace from eBPF with raw addresses
		ustack := stack(event.UserStack)
		// Print each event
		fmt.Printf("bin[%v] pid[%v]\n", event.taskComm(), pid)
		fmt.Println("  ADDRESS")
		fmt.Println("  ---------")
		for _, l := range ustack {
			fmt.Println(" ", l)
		}
		fmt.Println()
	}
}
