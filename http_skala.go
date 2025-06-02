package fasthttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type PrefetchedBytesInterface interface {
	BufIOReader() (io.Reader, error)
	IsPrefetchedBytesEmpty() bool
	SetPrefetchedBytes(r *bytes.Reader) error
}

func (resp *Response) BufIOReader() (io.Reader, error) {
	rs, err := getRequestStream(resp.bodyStream)
	if rs == nil {
		return nil, errors.Join(errors.New("nil response stream"), err)
	}
	if err != nil {
		return nil, err
	}
	return rs.reader, nil
}

func (resp *Response) IsPrefetchedBytesEmpty() bool {
	rs, err := getRequestStream(resp.bodyStream)
	if rs == nil || err != nil {
		return false
	}
	return rs.prefetchedBytes.Len() == 0 && rs.prefetchedBytes.Size() == 0
}

func (resp *Response) SetPrefetchedBytes(buf []byte) error {
	rs, err := getRequestStream(resp.bodyStream)
	if rs == nil {
		return errors.Join(errors.New("nil response stream"), err)
	}
	if err != nil {
		return err
	}
	rs.prefetchedBytes = bytes.NewReader(buf)

	return nil
}

func getRequestStream(reader io.Reader) (rs *requestStream, err error) {
	defer func() {
		if rvr := recover(); rvr != nil {
			err = fmt.Errorf("%v", rvr)
		}
	}()
	rs = ((*closeReader)(unsafe.Pointer(&reader)).
		Reader).(*closeReader).
		Reader.(*requestStream)
	return rs, err
}
