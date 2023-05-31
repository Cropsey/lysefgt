package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

func main() {
	var pid int
	flag.IntVar(&pid, "pid", 0, "PID of the profiled process")
	flag.Parse()
	log.SetOutput(os.Stderr)

	// Create a perf event array for the kernel to write perf records to.
	// These records will be read by userspace below.
	events, err := ebpf.NewMap(&ebpf.MapSpec{
		Type: ebpf.PerfEventArray,
		Name: "my_perf_array",
	})
	if err != nil {
		log.Fatalf("Creating perf event array: %s", err)
	}
	defer events.Close()

	// Open a perf reader from userspace into the perf event array created earlier.
	rd, err := perf.NewReader(events, os.Getpagesize())
	if err != nil {
		log.Fatalf("Creating event reader: %s", err)
	}
	defer rd.Close()

	// Metadata for the eBPF program used in this example.
	var progSpec = &ebpf.ProgramSpec{
		Name:    "my_perf_event_prog", // non-unique name, will appear in `bpftool prog list` while attached
		Type:    ebpf.PerfEvent,       // only PerfEvent programs can be attached to perf events
		License: "GPL",                // license must be GPL for calling kernel helpers like perf_event_output
	}

	// Minimal program that writes the static value '123' to the perf ring on
	// each event. Note that this program refers to the file descriptor of
	// the perf event array created above, which needs to be created prior to the
	// program being verified by and inserted into the kernel.
	progSpec.Instructions = asm.Instructions{
		// store the integer 123 at FP[-8]
		asm.Mov.Imm(asm.R2, 123),
		asm.StoreMem(asm.RFP, -8, asm.R2, asm.Word),

		// load registers with arguments for call of FnPerfEventOutput
		asm.LoadMapPtr(asm.R2, events.FD()), // file descriptor of the perf event array
		asm.LoadImm(asm.R3, 0xffffffff, asm.DWord),
		asm.Mov.Reg(asm.R4, asm.RFP),
		asm.Add.Imm(asm.R4, -8),
		asm.Mov.Imm(asm.R5, 4),

		// call FnPerfEventOutput, an eBPF kernel helper
		asm.FnPerfEventOutput.Call(),

		// set exit code to 0
		asm.Mov.Imm(asm.R0, 0),
		asm.Return(),
	}

	// Instantiate and insert the program into the kernel.
	prog, err := ebpf.NewProgram(progSpec)
	if err != nil {
		log.Fatalf("Creating ebpf program: %s", err)
	}
	defer prog.Close()
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
	if err := unix.IoctlSetInt(perfEventFD, unix.PERF_EVENT_IOC_SET_BPF, prog.FD()); err != nil {
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

		fmt.Printf("pid[%v]\n", pid)
		fmt.Println("  RECORD")
		fmt.Println("  ---------")
		fmt.Println(" ", record)
		fmt.Println()
	}
}
