package profile

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"syscall"
)

var cpu *os.File
var tr *os.File

// Start begins running CPU pprof and trace.
// it adds a Goroutine to listen for SIGINT and SIGTERM
// to call Stop. You should add a `defer Stop()`
// call to whichever function you call Start in.
func Start() {
	var err error
	cpu, err = os.Create("cpu.pprof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err = pprof.StartCPUProfile(cpu); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	tr, err = os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	if err = trace.Start(tr); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

		// Wait for a SIGINT or SIGKILL
		<-sigc
		Stop()
		os.Exit(0)
	}()
}

// Stop will cleanup the CPU pprof and trace, as well as do a memory profile dump
func Stop() {
	pprof.StopCPUProfile()
	trace.Stop()

	mem, err := os.Create("mem.pprof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(mem); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	if err := mem.Close(); err != nil {
		log.Fatalf("failed to close mem profile file: %v", err)
	}
	if err := tr.Close(); err != nil {
		log.Fatalf("failed to close trace file: %v", err)
	}
	log.Println("profiles saved")
}
