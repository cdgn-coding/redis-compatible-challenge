package resp

import (
	"bufio"
	"errors"
	"io"
)

var ErrBufferFull = errors.New("buffer full")

var ErrInvalidFormat = errors.New("invalid RESP format")

func ToBuffer(reader io.Reader, buffer []byte) (int, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	return scannerToBuffer(scanner, buffer)
}

func scannerToBuffer(scanner *bufio.Scanner, buffer []byte) (int, error) {
	if !scanner.Scan() {
		return 0, io.EOF
	}

	line := scanner.Bytes()
	if len(line) == 0 {
		return 0, ErrInvalidFormat
	}

	written := 0
	var err error

	switch line[0] {
	case '*': // Array
		written, err = readArray(scanner, buffer, line)
	case '$': // Bulk String
		written, err = readBulkString(scanner, buffer, line)
	case '+', '-', ':': // Simple String, Error, Integer
		written, err = appendLine(buffer, line)
	default:
		return 0, ErrInvalidFormat
	}

	if err != nil {
		return written, err
	}

	return written, nil
}

func readArray(scanner *bufio.Scanner, buffer []byte, firstLine []byte) (int, error) {
	written, err := appendLine(buffer, firstLine)
	if err != nil {
		return written, err
	}

	count := 0
	for i := 1; i < len(firstLine); i++ {
		if firstLine[i] == '\r' {
			break
		}
		count = count*10 + int(firstLine[i]-'0')
	}

	for i := 0; i < count; i++ {
		n, err := scannerToBuffer(scanner, buffer[written:])
		written += n
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func readBulkString(scanner *bufio.Scanner, buffer []byte, firstLine []byte) (int, error) {
	written, err := appendLine(buffer, firstLine)
	if err != nil {
		return written, err
	}

	length := 0
	for i := 1; i < len(firstLine); i++ {
		if firstLine[i] == '\r' {
			break
		}
		length = length*10 + int(firstLine[i]-'0')
	}

	if length < 0 {
		return written, nil // Null bulk string
	}

	for length > 0 {
		if !scanner.Scan() {
			return written, io.EOF
		}
		line := scanner.Bytes()
		n, err := appendLine(buffer[written:], line)
		written += n
		if err != nil {
			return written, err
		}
		length -= len(line)
	}

	return written, nil
}

func appendLine(buffer, line []byte) (int, error) {
	if len(buffer) < len(line)+2 {
		return 0, ErrBufferFull
	}
	copy(buffer, line)
	buffer[len(line)] = '\r'
	buffer[len(line)+1] = '\n'
	return len(line) + 2, nil
}
