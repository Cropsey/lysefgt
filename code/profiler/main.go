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
	"runtime"
	"time"

	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

// Use cilium/ebpf package to generate eBPF scaffold for programs and maps
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target native -type event bpf perf.c -- -I../headers

// recordBuffer forwards the perf.Reader records to a channel
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

// main is an entrypoint to the profiler
func main() {
	// Command line flag parsing
	var pid int
	var verbose int
	flag.IntVar(&pid, "pid", 0, "PID of the profiled process, when 0 it will try to profile all processes it has ability to do so")
	flag.IntVar(&verbose, "v", 0, "Verbosity of perf event logs, 0 prints only aggregate, 1 prints each perf event record")
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

	// Wrap perf reader in a goroutine with buffer
	recordsChan := make(chan *perf.Record, 1024)
	go recordBuffer(rd, recordsChan)

	log.Println("Waiting for events..")

	var perfEventFDs []int
	if pid == 0 { // Profile all processes on all CPUs
		for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
			log.Printf("Setting all process profiler on CPU %v", cpu)
			// This requires CAP_PERFMON or CAP_SYS_ADMIN or /proc/sys/kernel/perf_event_paranoid
			anyPid := -1      // Profile all processes and threads on the specified CPU
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
				anyPid,                    // Profile all processes
				cpu,                       // The CPU to profile all processes
				groupLeader,               // Configure event grouping, -1 means disabled
				unix.PERF_FLAG_FD_CLOEXEC, // Close the event fd on execve, avoiding possible race conditions
			)
			if err != nil {
				log.Fatalf("Failed to create the perf event: %v", err)
			}
			perfEventFDs = append(perfEventFDs, perfEventFD)
		}
	} else { // Running with specific PID to profile
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
		perfEventFDs = append(perfEventFDs, perfEventFD)
	}

	for _, perfEventFD := range perfEventFDs {
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
	}

	// Create ELF parser for the binary running as process with PID
	elf := newElf(pid)

	// Stack trace aggregator to print stats summary at the end
	stats := newStats()
	defer stats.summary()

	// Capture SIGINT for graceful termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Forward declarations for the loop variables
	var event bpfEvent
	var record perf.Record

	// Loop forever and process stack traces
	for {
		// Select whether there is a new record or SIGTERM to close the program
		select {
		case <-c: // SIGTERM -> graceful termination
			return
		case r := <-recordsChan: // Record from perf event fd reader
			if r == nil {
				// unrecoverable error from perfFD
				return
			}
			record = *r
		}

		// Read eBPF binary data into go native structure, the ABI compatibility is ensured by cilium/ebpf
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			log.Printf("parsing perf event failed: %s", err)
		}

		// Transform the stack trace from eBPF with raw addresses to something more readable
		ustack := elf.humanReadableStack(int(event.Pid), event.UserStack)
		if ustack == nil {
			continue
		}

		// Aggregate stack trace statistics
		for _, l := range ustack {
			stats.add(int(event.Pid), l)
		}

		if verbose >= 1 {
			// Print each event
			fmt.Printf("bin[%v] pid[%v]\n", event.taskComm(), event.Pid)
			fmt.Println("  ADDRESS    PC         SYMBOL                             FILE:LINE")
			fmt.Println("  ---------  ---------  ---------------------------------  ------------------------------------")
			for _, l := range ustack {
				fmt.Println(" ", l)
			}
			fmt.Println()
		}
	}
}
