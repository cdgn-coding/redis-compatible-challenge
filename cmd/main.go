package main

import (
	"context"
	"flag"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/server"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
)

var port = flag.String("port", "3000", "port")
var threads = flag.Int("threads", 1, "number of threads")
var cpuProfile = flag.Bool("cpuprofile", false, "profile cpu")
var memProfile = flag.Bool("memprofile", false, "profile memory")
var mutexProfile = flag.Bool("mutexprofile", false, "profile mutexes")
var reload = flag.Bool("reload", true, "reload memory")
var memfile = flag.String("memfile", "memory.resp", "path to memory file")
var global = flag.Bool("global", false, "use global path")

func main() {
	flag.Parse()

	var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Configure concurrency
	if *threads > runtime.NumCPU() {
		logger.Println("Warning max threads is beyond number of cores")
	}
	logger.Printf("Using %d CPU", *threads)
	runtime.GOMAXPROCS(*threads)

	// Configure profiler
	if *cpuProfile {
		logger.Println("Starting CPU profiler")
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		if pprof.StartCPUProfile(f) != nil {
			logger.Fatal(err)
			return
		}
		defer pprof.StopCPUProfile()
	}
	// Enable mutex profiling
	if *mutexProfile {
		runtime.SetMutexProfileFraction(1)
		defer runtime.SetMutexProfileFraction(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan struct{})

	opts := engine.EngineOptions{
		Load:       reload,
		GlobalPath: global,
	}
	if *memfile != "" {
		opts.File = memfile
	}
	eng, err := engine.NewEngine(opts)
	if err != nil {
		logger.Fatalf("Error creating engine: %v", err)
	}
	serv := server.NewServer(eng, logger)
	go serv.StartServer(ctx, *port, ready)
	<-ready

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
	cancel()

	if *memProfile {
		f, err := os.Create("mem.prof")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err = pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
