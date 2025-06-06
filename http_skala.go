package fasthttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

var (
	ErrResponseIsNil             = errors.New("response is nil")
	ErrResponseStreamIsNil       = errors.New("response stream is nil")
	ErrPrefetchedBytesIsNotEmpty = errors.New("prefetched bytes is not empty")
)

// FirstBodyBytes returns firstBytesBodySize from response body.
//
// Method FirstBodyBytes is idempotent for this it uses prefetchedBytes from requestStream.
func (resp *Response) FirstBodyBytes(firstBytesBodySize int) ([]byte, error) {
	if resp == nil {
		return nil, ErrResponseIsNil
	}
	if resp.isNotEmptyPrefetchedBytes() {
		return nil, ErrPrefetchedBytesIsNotEmpty
	}
	prefetchedBytes := make([]byte, firstBytesBodySize)

	bioReader, errBufIOReader := resp.bufIOReader()
	if errBufIOReader != nil {
		return nil, errBufIOReader
	}
	_, errRead := bioReader.Read(prefetchedBytes)
	if errRead != nil {
		return nil, errRead
	}
	errSetPrefetchedBytes := resp.setPrefetchedBytes(prefetchedBytes)
	if errSetPrefetchedBytes != nil {
		return nil, errSetPrefetchedBytes
	}
	return prefetchedBytes, nil
}

func (resp *Response) bufIOReader() (io.Reader, error) {
	if resp == nil {
		return nil, ErrResponseIsNil
	}
	rs, err := castRequestStream(resp.bodyStream)
	if rs == nil {
		return nil, errors.Join(ErrResponseStreamIsNil, err)
	}
	if err != nil {
		return nil, err
	}
	return rs.reader, nil
}

func (resp *Response) isNotEmptyPrefetchedBytes() bool {
	if resp == nil {
		return true
	}
	rs, err := castRequestStream(resp.bodyStream)
	if rs == nil || err != nil {
		return true
	}
	return rs.prefetchedBytes.Len() != 0 || rs.prefetchedBytes.Size() != 0
}

func (resp *Response) setPrefetchedBytes(buf []byte) error {
	if resp == nil {
		return ErrResponseIsNil
	}
	rs, err := castRequestStream(resp.bodyStream)
	if rs == nil {
		return errors.Join(ErrResponseStreamIsNil, err)
	}
	if err != nil {
		return err
	}
	rs.prefetchedBytes = bytes.NewReader(buf)

	return nil
}

func castRequestStream(reader io.Reader) (rs *requestStream, err error) {
	defer func() {
		if rvr := recover(); rvr != nil {
			err = fmt.Errorf("fail when get request stream: %v", rvr)
		}
	}()
	rs = ((*closeReader)(unsafe.Pointer(&reader)).
		Reader).(*closeReader).
		Reader.(*requestStream)
	return rs, err
}
