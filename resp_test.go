package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRespSerializer_SerializeString(t *testing.T) {
	cases := []struct {
		data     string
		expected []byte
		err      error
		message  string
	}{
		{
			data:     "OK",
			expected: []byte("+OK\r\n"),
			message:  "Simple string",
			err:      nil,
		},
		{
			data:     "data with \r\n",
			expected: nil,
			message:  "Incorrect simple string",
			err:      SimpleStringsError,
		},
	}

	serializer := RespSerializer{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			actual, err := serializer.SerializeString(c.data)
			if !errors.Is(err, c.err) || bytes.Compare(actual, c.expected) != 0 {
				t.Errorf("Actual data %s, actual error %v, expected data %s, expected error %v", actual, err, c.expected, c.err)
			}
		})
	}
}

func TestRespSerializer_SerializeInteger(t *testing.T) {
	cases := []struct {
		data     int64
		expected []byte
		err      error
		message  string
	}{
		{
			data:     127,
			expected: []byte(":127\r\n"),
			message:  "Positive integer",
			err:      nil,
		},
	}

	serializer := RespSerializer{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			actual, err := serializer.SerializeInteger(c.data)
			if !errors.Is(err, c.err) || bytes.Compare(actual, c.expected) != 0 {
				t.Errorf("Actual data %s, actual error %v, expected data %s, expected error %v", actual, err, c.expected, c.err)
			}
		})
	}
}

func TestRespSerializer_SerializeError(t *testing.T) {
	cases := []struct {
		data     error
		expected []byte
		err      error
		message  string
	}{
		{
			data:     errors.New("unknown error"),
			expected: []byte("-unknown error\r\n"),
			message:  "Non empty error",
			err:      nil,
		},
		{
			data:     nil,
			expected: nil,
			message:  "Empty error",
			err:      EmptyError,
		},
	}

	serializer := RespSerializer{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			actual, err := serializer.SerializeError(c.data)
			if !errors.Is(err, c.err) || bytes.Compare(actual, c.expected) != 0 {
				t.Errorf("Actual data %s, actual error %v, expected data %s, expected error %v", actual, err, c.expected, c.err)
			}
		})
	}
}

func TestRespSerializer_SerializeBulkString(t *testing.T) {
	cases := []struct {
		data     string
		expected []byte
		err      error
		message  string
	}{
		{
			data:     "normal string",
			expected: []byte("$13\r\nnormal string\r\n"),
			message:  "Bulk string with no especial characters",
			err:      nil,
		},
		{
			data:     "string with\nbreak line",
			expected: []byte("$22\r\nstring with\nbreak line\r\n"),
			message:  "Bulk string with break line",
			err:      nil,
		},
		{
			data:     "",
			expected: []byte("$0\r\n\r\n"),
			message:  "Empty Bulk string",
			err:      nil,
		},
	}

	serializer := RespSerializer{}
	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			actual, err := serializer.SerializeBulkString(c.data)
			if !errors.Is(err, c.err) || bytes.Compare(actual, c.expected) != 0 {
				t.Errorf("Actual data %s, actual error %v, expected data %s, expected error %v", actual, err, c.expected, c.err)
			}
		})
	}
}
