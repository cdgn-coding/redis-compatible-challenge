package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/cdgn-coding/redis-compatible-challenge/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/resp"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

var eng = engine.NewEngine()

func handleClient(conn net.Conn) {
	var parser = resp.RespParser{}
	var serializer = resp.RespSerializer{}
	var serialized *bytes.Buffer
	scanner := bufio.NewScanner(conn)
	for {
		payload, err := parser.ParseWithScanner(scanner)

		if err != nil {
			logger.Printf("client closed connection from %s", conn.RemoteAddr())
			return
		}

		// Process payload
		res, err := eng.Process(payload)

		// Report engine errors
		if err != nil {
			logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized.Bytes())
			return
		}

		// Serialize response
		serialized, err = serializer.Serialize(res)

		// Report serialization errors
		if err != nil {
			logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized.Bytes())
			continue
		}

		// Write response
		_, err = conn.Write(serialized.Bytes())
		if err != nil {
			logger.Println(err)
			return
		}
	}

	defer conn.Close()
}

func startServer(ctx context.Context, port string) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal(err)
	}

	defer listener.Close()

	logger.Println("Listening on :3000...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				logger.Println(err)
				continue
			}

			logger.Printf("accepted connection from %s", conn.RemoteAddr())
			go handleClient(conn)
		}
	}
}

var port = flag.String("port", "3000", "redis port")
var cpu = flag.Int("cpu", 1, "cpu cores")
var profile = flag.Bool("profile", false, "profile program")

func main() {
	flag.Parse()

	if *cpu > runtime.NumCPU() {
		logger.Println("Warning cpu is beyond number of cores")
	}
	logger.Printf("Using %d CPU", *cpu)
	runtime.GOMAXPROCS(*cpu)

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
	go startServer(ctx, *port)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
	cancel()
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
