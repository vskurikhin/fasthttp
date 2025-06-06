package fasthttp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func Test_Response_isNotEmptyPrefetchedBytes(t *testing.T) {
	type args struct {
		response *Response
		buf      []byte
	}
	type want struct {
		result bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"negative #1",
			args{
				(*Response)(nil),
				nil,
			},
			want{true},
		},
		{
			"negative #2",
			args{
				&Response{},
				nil,
			},
			want{true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.args.response.isNotEmptyPrefetchedBytes()
			if tt.want.result != result {
				t.Errorf("got: %v, want: %v", result, tt.want.result)
			}
		})
	}
}

func Test_Response_setPrefetchedBytes(t *testing.T) {
	type args struct {
		response *Response
		buf      []byte
	}
	type want struct {
		err         error
		errContains string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"negative #1",
			args{
				(*Response)(nil),
				nil,
			},
			want{err: ErrResponseIsNil},
		},
		{
			"negative #2",
			args{
				&Response{},
				nil,
			},
			want{errContains: "response stream is nil"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.response.setPrefetchedBytes(tt.args.buf)
			if tt.want.err != nil && !errors.Is(err, tt.want.err) {
				t.Errorf("got: %v, want: %v", err, tt.want.err)
			}
			if tt.want.errContains != "" {
				if err != nil && !strings.Contains(err.Error(), tt.want.errContains) {
					t.Errorf("err got = %v, want = %v", err.Error(), tt.want.errContains)
				}
			}
		})
	}
}

func Test_castRequestStream(t *testing.T) {
	resp := &Response{}
	bodyBuf := resp.bodyBuffer()
	bodyBuf.Reset()
	r := bufio.NewReader(bytes.NewReader([]byte{}))
	test1RequestStream := acquireRequestStream(bodyBuf, r, &resp.Header)
	test1CloseReade := newCloseReaderWithError(test1RequestStream, func(wErr error) error {
		return nil
	})
	type args struct {
		reader io.Reader
	}
	type want struct {
		rs          *requestStream
		errContains string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"positive #1",
			args{test1CloseReade},
			want{rs: test1RequestStream, errContains: ""},
		},
		{
			"negative #1",
			args{
				bytes.NewBuffer([]byte{}),
			},
			want{rs: nil, errContains: "fail when get request stream:"},
		},
		{
			"negative #2",
			args{
				nil,
			},
			want{rs: nil, errContains: "fail when get request stream: interface conversion: io.Reader is nil"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := castRequestStream(tt.args.reader)
			if rs != tt.want.rs {
				t.Errorf("requestStream got = %v, want %v", rs, tt.want.rs)
			}
			if tt.want.errContains != "" {
				if err != nil && !strings.Contains(err.Error(), tt.want.errContains) {
					t.Errorf("err got = %v, want = %v", err.Error(), tt.want.errContains)
				}
			}
		})
	}
}
