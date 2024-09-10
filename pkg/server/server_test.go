package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/resp"
	"github.com/stretchr/testify/suite"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	serv   *Server
	cancel context.CancelFunc
}

func (suite *TestSuite) SetupSuite() {
	eng, _ := engine.NewEngine(engine.EngineOptions{})
	serv := &Server{
		eng:    eng,
		logger: log.New(io.Discard, "", log.LstdFlags),
	}
	ctx, cancel := context.WithCancel(context.Background())

	suite.serv = serv
	suite.cancel = cancel

	ready := make(chan struct{})
	go serv.StartServer(ctx, "3000", ready)
	<-ready
}

func (suite *TestSuite) TearDownSuite() {
	suite.cancel()
}

func (suite *TestSuite) TestServer_SET_GET() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		suite.T().Fatal(err)
	}

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	conn.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))
	<-time.After(1 * time.Second)

	parser := resp.RespParser{}

	res, err := parser.ParseScanner(scanner)
	if err != nil {
		suite.T().Fatal(err)
	}

	if reflect.TypeOf(res).Kind() != reflect.String || res.(string) != "OK" {
		suite.T().Fatal("Expected SET to return string OK")
	}

	conn.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"))
	<-time.After(1 * time.Second)

	res, err = parser.ParseScanner(scanner)
	if err != nil {
		suite.T().Fatal(err)
	}

	if reflect.TypeOf(res).Kind() != reflect.String || res.(string) != "value" {
		suite.T().Fatal("Expected GET key to return string value")
	}
}

func (suite *TestSuite) TestConcurrencyServer() {
	var numGoroutines = 10
	var numRequests = 1000
	var errs = make([]error, numGoroutines)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", ":3000")
			if err != nil {
				errs[i] = fmt.Errorf("connection error: %v", err)
				return
			}
			defer conn.Close()

			parser := resp.RespParser{}
			ticker := time.NewTicker(1 * time.Millisecond)
			defer ticker.Stop()

			scanner := bufio.NewScanner(conn)
			scanner.Split(bufio.ScanLines)

			timeout := time.After(10 * time.Minute)

			for j := 0; j < numRequests; {
				// Write request
				_, err = conn.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))
				if err != nil {
					errs[i] = fmt.Errorf("write error: %v", err)
					return
				}
				select {
				case <-timeout:
					errs[i] = fmt.Errorf("global timeout reached")
					return

				case <-ticker.C:
					_ = conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
					_, err = parser.ParseScanner(scanner)

					if err != nil {
						continue
					}

					j++
				}
			}
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			suite.T().Logf("Error in goroutine %d: %v", i, err)
			suite.T().Fail()
		}
	}
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
