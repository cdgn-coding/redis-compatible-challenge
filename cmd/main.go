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
var profile = flag.Bool("profile", false, "profile program")

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
	if *profile {
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

	ctx, cancel := context.WithCancel(context.Background())

	eng := engine.NewEngine()
	serv := server.NewServer(eng, logger)
	go serv.StartServer(ctx, *port)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
	cancel()
}
