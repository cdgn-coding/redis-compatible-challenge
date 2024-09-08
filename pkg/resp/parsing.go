package resp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

type RespParser struct{}

var CannotReadDataError = errors.New("cannot read data")

var TypeMismatchError = errors.New("type mismatch")

var UnsupportedType = errors.New("unsupported type")

var NumberOfBytesOff = errors.New("number of bytes off")

func (p RespParser) Parse(data []byte) (interface{}, error) {
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	return p.ParseWithScanner(scanner)
}

func (p RespParser) ParseWithReader(reader io.Reader) (interface{}, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	return p.ParseWithScanner(scanner)
}

func (p RespParser) ParseWithScanner(scanner *bufio.Scanner) (interface{}, error) {
	var line []byte
	var i, count int64
	var read, totalBytes int
	var err error
	var part interface{}

	for scanner.Scan() {
		line = scanner.Bytes()

		switch line[0] {
		case '*':
			count, err = strconv.ParseInt(string(line[1:]), 10, 64)
			if err != nil {
				return nil, errors.Join(err, TypeMismatchError)
			}

			result := make([]interface{}, count)
			for i = 0; i < count; i++ {
				part, err = p.ParseWithScanner(scanner)

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

	return nil, CannotReadDataError
}
