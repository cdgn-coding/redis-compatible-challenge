# Redis Challenge Implementation

This project is an implementation of a Redis-like database server, based on the challenge from [Coding Challenges](https://codingchallenges.fyi/challenges/challenge-redis).

This implementation is **dependency free**.

## Project Overview

The goal is to create a simplified version of Redis, implementing core functionality and following the Redis protocol specification.

## Protocol Implementation Checklist

Based on the [Redis Protocol Specification](https://redis.io/docs/latest/develop/reference/protocol-spec/), implement the following aspects:

- [x] RESP2 (Redis Serialization Protocol) parsing
- [x] RESP2 (Redis Serialization Protocol) serialization
- [x] Client-server communication
- [x] Implement commands
  - [x] PING
  - [x] ECHO
  - [x] SET
  - [x] GET
  - [x] DEL
  - [x] EXISTS
  - [x] INCR
  - [x] DECR
  - [x] LPUSH
  - [x] RPUSH
  - [x] SAVE

## Benchmark

Implementation Benchmark

```
cdgn@MBP-de-Carlos ~ % redis-benchmark -p 3000 -t INCR,GET,SET,LPUSH,RPUSH -q
WARNING: Could not fetch server CONFIG
SET: 66489.37 requests per second, p50=0.367 msec                   
GET: 64977.26 requests per second, p50=0.375 msec                   
INCR: 69735.01 requests per second, p50=0.359 msec                   
LPUSH: 57077.62 requests per second, p50=0.407 msec                   
RPUSH: 52164.84 requests per second, p50=0.431 msec                            
```

Redis benchmark in the same machine

```
cdgn@MBP-de-Carlos ~ % redis-benchmark -p 6379 -t INCR,GET,SET,LPUSH,RPUSH -q
SET: 88028.16 requests per second, p50=0.119 msec                   
GET: 87950.75 requests per second, p50=0.119 msec                   
INCR: 85251.49 requests per second, p50=0.119 msec                   
LPUSH: 87873.46 requests per second, p50=0.119 msec                   
RPUSH: 89445.44 requests per second, p50=0.119 msec
```

Machine details

```
[cdgn@MBP-de-Carlos ~ % sysctl -a | grep cpu
hw.cpufamily: 943936839
machdep.cpu.brand_string: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
machdep.cpu.core_count: 4
machdep.cpu.thread_count: 8

cdgn@MBP-de-Carlos ~ % sysctl hw.memsize
hw.memsize: 17179869184
```

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/cdgn-coding/redis-compatible-challenge
   cd redis-compatible-challenge
   ```

2. Ensure you have Go installed on your system. You can download it from [golang.org](https://golang.org/).

3. Build the program:
   ```
   go build -o ./redis-server ./cmd/main.go
   ```

4. Run the server:
   ```
   ./redis-server
   ```

You can start the server with the following command line options:

* port: Set the server port (default: 3000)
* threads: Specify the number of threads to use (default: 1)
* cpuprofile: Enable CPU profiling (default: false)
* memprofile: Enable memory profiling (default: false)
* mutexprofile: Enable mutex profiling (default: false)
* reload: Enable reloading of memory from file on startup (default: true)
* memfile: Specify the path to the memory file (default: "memory.resp")
* global: Use a global path for configuration and data (default: false)

Here's an example command to run the server on port 8000, with CPU and memory profiling enabled, and using 4 threads:

```
./redis-compatible-challenge -port="8000" -threads=4 -cpuprofile=true -memprofile=true -mutexprofile=true -reload=true -memfile="path/to/your/memory.resp" -global=false
```

## Testing

To run the tests for this project:

1. Navigate to the project directory:
   ```
   cd redis-challenge
   ```

2. Run the Go test command:
   ```
   go test ./...
   ```

This will run all tests in the project and its subdirectories.

To run tests with verbose output:
```
go test -v ./...
```

To run tests with race condition detection

```
go test -race ./...
```

## Linting

```
golangci-lint run ./...
gosec ./...
govulncheck ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License:

MIT License

Copyright (c) 2024 Carlos David Gonzalez Nexans

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.