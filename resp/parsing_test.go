package resp

import (
	"errors"
	"reflect"
	"testing"
)

func TestRespParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    interface{}
		wantErr bool
	}{
		{
			name:    "Simple integer",
			data:    []byte(":127\r\n"),
			want:    int64(127),
			wantErr: false,
		},
		{
			name:    "Simple string",
			data:    []byte("+OK\r\n"),
			want:    "OK",
			wantErr: false,
		},
		{
			name:    "Simple error",
			data:    []byte("-some error"),
			want:    errors.New("some error"),
			wantErr: false,
		},
		{
			want:    "normal string",
			data:    []byte("$13\r\nnormal string\r\n"),
			name:    "Bulk string with no especial characters",
			wantErr: false,
		},
		{
			want:    "string with\nbreak line",
			data:    []byte("$22\r\nstring with\nbreak line\r\n"),
			name:    "Bulk string with break line",
			wantErr: false,
		},
		{
			want:    nil,
			data:    []byte("$3000\r\nstring with\nbreak line\r\n"),
			name:    "Bulk string bad number of bytes",
			wantErr: true,
		},
		{
			want:    []interface{}{},
			data:    []byte("*0\r\n"),
			name:    "Empty array",
			wantErr: false,
		},
		{
			want: []interface{}{
				"hello",
				"world",
			},
			data:    []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			name:    "Array two bulk strings",
			wantErr: false,
		},
		{
			want:    []interface{}{int64(1), int64(2), int64(3)},
			data:    []byte("*3\r\n:1\r\n:2\r\n:3\r\n"),
			name:    "Array of three integers",
			wantErr: false,
		},
		{
			want:    []interface{}{int64(1), int64(2), int64(3), int64(4), "hello"},
			data:    []byte("*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n"),
			name:    "Array of mixed data types",
			wantErr: false,
		},
		{
			want:    []interface{}{[]interface{}{int64(1), int64(2), int64(3), int64(4)}, []interface{}{"hello", "world"}},
			data:    []byte("*2\r\n*4\r\n:1\r\n:2\r\n:3\r\n:4\r\n*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			name:    "Array of arrays",
			wantErr: false,
		},
		{
			want:    []interface{}{errors.New("some error")},
			data:    []byte("*1\r\n-some error\r\n"),
			name:    "Array with error",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := RespParser{}
			got, err := p.Parse(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
