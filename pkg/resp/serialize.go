package resp

import (
	"bytes"
	"container/list"
	"errors"
	"github.com/cdgn-coding/redis-compatible-challenge/pkg/concurrency"
	"iter"
	"reflect"
	"strconv"
	"sync"
)

var listType = reflect.TypeOf(&list.List{})

var EmptyError = errors.New("cannot serialize empty error")

var SimpleStringsError = errors.New("cannot serialize simple strings with \\n or \\r")

var UnknownType = errors.New("cannot serialize unknown type")

var ArrayError = errors.New("cannot serialize array")

var RespNull = "$-1\r\n"

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

type RespSerializer struct{}

func (s RespSerializer) Serialize(element interface{}) (*bytes.Buffer, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	err := s.SerializeWithBuffer(buf, element)
	return buf, err
}

func (s RespSerializer) SerializeWithBuffer(buf *bytes.Buffer, element interface{}) error {
	if element == nil {
		buf.WriteString(RespNull)
		return nil
	}

	t := reflect.TypeOf(element)

	var err error
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		err = s.SerializeArray(buf, element.([]interface{}))
	case reflect.Int:
		err = s.SerializeInteger(buf, int64(element.(int)))
	case reflect.Int64:
		err = s.SerializeInteger(buf, element.(int64))
	case reflect.String:
		err = s.SerializeBulkString(buf, element.(string))
	case reflect.Ptr:
		if t.Implements(errorInterface) {
			err = s.SerializeError(buf, element.(error))
		} else if t.AssignableTo(concurrency.ConcurrentListType) {
			err = s.SerializeIterable(buf, element.(*concurrency.ConcurrentList).Iterator())
		} else if t.AssignableTo(listType) {
			err = s.SerializeIterable(buf, s.collectList(element.(*list.List)))
		} else {
			return UnknownType
		}
	default:
		return UnknownType
	}

	if err != nil {
		bufferPool.Put(buf)
		return err
	}

	return nil
}

func (RespSerializer) collectList(list *list.List) iter.Seq[interface{}] {
	return func(yield func(interface{}) bool) {
		for e := list.Front(); e != nil; e = e.Next() {
			if !yield(e.Value) {
				return
			}
		}
	}
}

func (RespSerializer) SerializeString(buf *bytes.Buffer, data string) error {
	for _, c := range data {
		if c == '\n' || c == '\r' {
			return SimpleStringsError
		}
	}

	buf.WriteByte('+')
	buf.WriteString(data)
	buf.WriteString("\r\n")
	return nil
}

func (RespSerializer) SerializeInteger(buf *bytes.Buffer, data int64) error {
	buf.WriteByte(':')
	buf.WriteString(strconv.FormatInt(data, 10))
	buf.WriteString("\r\n")
	return nil
}

func (RespSerializer) SerializeError(buf *bytes.Buffer, data error) error {
	if data == nil {
		return EmptyError
	}
	buf.WriteByte('-')
	buf.WriteString(data.Error())
	buf.WriteString("\r\n")
	return nil
}

func (RespSerializer) SerializeBulkString(buf *bytes.Buffer, data string) error {
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(len(data)))
	buf.WriteString("\r\n")
	buf.WriteString(data)
	buf.WriteString("\r\n")
	return nil
}

func (s RespSerializer) SerializeArray(buf *bytes.Buffer, data []interface{}) error {
	buf.WriteByte('*')
	buf.WriteString(strconv.Itoa(len(data)))
	buf.WriteString("\r\n")
	var err error

	tempBuf := bytes.Buffer{}
	for _, element := range data {
		tempBuf.Reset()
		err = s.SerializeWithBuffer(&tempBuf, element)

		if err != nil {
			return errors.Join(err, ArrayError)
		}

		buf.Write(tempBuf.Bytes())
	}

	return nil
}

func (s RespSerializer) SerializeIterable(buf *bytes.Buffer, data iter.Seq[interface{}]) error {
	var err error
	tempBuf := bytes.Buffer{}

	var count = 0
	for element := range data {
		err = s.SerializeWithBuffer(&tempBuf, element)
		count++

		if err != nil {
			return errors.Join(err, ArrayError)
		}
	}

	buf.WriteByte('*')
	buf.WriteString(strconv.Itoa(count))
	buf.WriteString("\r\n")
	buf.Write(tempBuf.Bytes())

	return nil
}
