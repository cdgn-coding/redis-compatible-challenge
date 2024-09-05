# Redis Challenge Implementation

This project is an implementation of a Redis-like database server, based on the challenge from [Coding Challenges](https://codingchallenges.fyi/challenges/challenge-redis).

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
  - [ ] SET
    - [x] Basic
    - [ ] Options
  - [x] GET
  - [x] DEL
  - [x] EXISTS
  - [x] INCR
  - [x] DECR
  - [x] LPUSH
  - [x] RPUSH
  - [ ] SAVE

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/cdgn-coding/redis-compatible-challenge
   cd redis-compatible-challenge
   ```

2. Ensure you have Go installed on your system. You can download it from [golang.org](https://golang.org/).

3. Build the program:
   ```
   go build -o redis-server
   ```

4. Run the server:
   ```
   ./redis-server
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

## Benchmark

Implementation Benchmark

```
cdgn@MBP-de-Carlos ~ % redis-benchmark -p 3000 -t INCR,GET,SET,LPUSH,RPUSH -q
WARNING: Could not fetch server CONFIG
SET: 71428.57 requests per second, p50=0.335 msec                   
GET: 61050.06 requests per second, p50=0.351 msec                   
INCR: 73909.83 requests per second, p50=0.335 msec                   
LPUSH: 58105.75 requests per second, p50=0.391 msec                   
RPUSH: 54171.18 requests per second, p50=0.375 msec                   
```

Redis benchmark in the same machine

```
cdgn@MBP-de-Carlos ~ % redis-benchmark -p 6379 -t INCR,GET,SET,LPUSH,RPUSH -q
SET: 26281.21 requests per second, p50=1.631 msec                   
GET: 28208.74 requests per second, p50=1.551 msec                   
INCR: 26219.19 requests per second, p50=1.639 msec                   
LPUSH: 21706.10 requests per second, p50=1.999 msec                   
RPUSH: 27624.31 requests per second, p50=1.567 msec
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