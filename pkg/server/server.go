package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/engine"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/resp"
	"log"
	"net"
)

type Server struct {
	eng    *engine.Engine
	logger *log.Logger
}

func NewServer(eng *engine.Engine, logger *log.Logger) *Server {
	return &Server{
		eng:    eng,
		logger: logger,
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	var parser = resp.RespParser{}
	var serializer = resp.RespSerializer{}
	var serialized *bytes.Buffer
	var scanner = parser.CreateScanner(conn)

	for {
		payload, err := parser.ParseScanner(scanner)

		if err != nil {
			s.logger.Printf("client closed connection from %s", conn.RemoteAddr())
			return
		}

		// Process payload
		res, err := s.eng.Process(payload)

		// Report engine errors
		if err != nil {
			s.logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized.Bytes())
			continue
		}

		// Serialize response
		serialized, err = serializer.Serialize(res)

		// Report serialization errors
		if err != nil {
			s.logger.Println(err)
			serialized, _ = serializer.Serialize(err)
			_, err = conn.Write(serialized.Bytes())
			continue
		}

		// Write response
		_, err = conn.Write(serialized.Bytes())
		if err != nil {
			s.logger.Println(err)
			return
		}
	}
}

func (s *Server) StartServer(ctx context.Context, port string, ready chan struct{}) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		s.logger.Fatal(err)
	}

	defer listener.Close()

	s.logger.Printf("Listening on %s", listener.Addr())
	ready <- struct{}{}
	defer close(ready)

	for {
		select {
		case <-ctx.Done():
			err = listener.Close()
			if err != nil {
				s.logger.Printf("Error closing listener: %s", err)
				return
			}
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.logger.Println(err)
				continue
			}

			s.logger.Printf("Accepted connection from %s", conn.RemoteAddr())
			go s.handleClient(conn)
		}
	}
}
