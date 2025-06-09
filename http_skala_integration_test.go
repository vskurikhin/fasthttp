package fasthttp_test

import (
	"net"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestClientGetWithBodyStreamingAndFirstBodyBytes(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			body := ctx.Request.Body()
			ctx.Write(body) //nolint:errcheck
		},
	}
	go s.Serve(ln) //nolint:errcheck
	c := &fasthttp.Client{
		MaxResponseBodySize: 16,
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://example.com")
	body := "test" + strings.Repeat("a", 16380)
	req.SetBodyString(body)
	res.StreamBody = true
	err := c.Do(req, res)
	if err != nil {
		t.Fatal(err)
	}
	first, err := res.FirstBodyBytes(4)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != "test" {
		t.Fatalf("first is not equal to test: %s", first)
	}
	if len(res.Body()) == 0 {
		t.Fatal("missing request body")
	}
	if len(res.Body()) != 16384 {
		t.Fatal("unexpected response body size")
	}
	firstFromBody := string(res.Body())[:4]
	if firstFromBody != "test" {
		t.Fatalf("first body is not equal to test: %s", firstFromBody)
	}
	if string(res.Body()) != body {
		t.Fatalf("body is not equal to test: %s", body)
	}
}
