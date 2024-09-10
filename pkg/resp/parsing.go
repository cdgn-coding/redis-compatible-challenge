package resp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"iter"
	"strconv"
)

type RespParser struct{}

var CannotReadDataError = errors.New("cannot read data")

var TypeMismatchError = errors.New("type mismatch")

var UnsupportedType = errors.New("unsupported type")

var NumberOfBytesOff = errors.New("number of bytes off")

var EmptyPayload = errors.New("empty")

func (p RespParser) Parse(data []byte) (interface{}, error) {
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	return p.ParseScanner(scanner)
}

func (p RespParser) CreateScanner(reader io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	return scanner
}

func (p RespParser) ParseScanner(scanner *bufio.Scanner) (interface{}, error) {
	var line []byte
	var i, count int64
	var read, totalBytes int
	var err error
	var part interface{}

	if !scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}

		return nil, io.EOF
	}

	line = scanner.Bytes()

	switch line[0] {
	case '*':
		count, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, errors.Join(err, TypeMismatchError)
		}

		result := make([]interface{}, count)
		for i = 0; i < count; i++ {
			part, err = p.ParseScanner(scanner)

			if err != nil {
				return nil, err
			}

			result[i] = part
		}
		return result, nil
	case ':':
		i, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, errors.Join(err, TypeMismatchError)
		}

		return i, nil
	case '+':
		return string(line[1:]), nil
	case '-':
		return errors.New(string(line[1:])), nil
	case '$':
		totalBytes, err = strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, errors.Join(err, CannotReadDataError)
		}

		read = 0
		result := make([]byte, totalBytes)
		for read < totalBytes {
			scanner.Scan()
			current := scanner.Bytes()

			if current == nil {
				return nil, errors.Join(CannotReadDataError, NumberOfBytesOff)
			}

			copy(result[read:read+len(current)], current)
			read += len(current)

			if read < totalBytes {
				result[read] = '\n'
				read += 1
			}
		}

		return string(result), nil
	default:
		return nil, UnsupportedType
	}
}

type ParseResult struct {
	value interface{}
	err   error
}

func (p ParseResult) Value() interface{} {
	return p.value
}

func (p ParseResult) Err() error {
	return p.err
}

func NewParseResult(value interface{}, err error) ParseResult {
	return ParseResult{
		value: value,
		err:   err,
	}
}

func (p RespParser) Iterate(scanner *bufio.Scanner) iter.Seq[ParseResult] {
	return func(yield func(ParseResult) bool) {
		for {
			data, err := p.ParseScanner(scanner)

			if errors.Is(err, io.EOF) {
				return
			}

			if !yield(NewParseResult(data, err)) {
				return
			}
		}
	}
}
