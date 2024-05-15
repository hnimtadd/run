package sdk

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net/http"
	"os"

	pb "github.com/hnimtadd/run/pbs/gopb/v1"
	"google.golang.org/protobuf/proto"
)

type responseWriter struct {
	header http.Header
	buffer *bytes.Buffer
	code   int
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		header: http.Header{},
		buffer: new(bytes.Buffer),
	}
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	return w.buffer.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.code = status
}

// Handle unmarshal the marshaled request and process
func Handle(h http.Handler) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("sdk: cannot read stdin, err: %v", err)
	}

	req := new(pb.HTTPRequest)
	if err := proto.Unmarshal(b, req); err != nil {
		log.Fatalf("sdk: cannot unmarshal the proto request, err: %v", err)
	}

	w := newResponseWriter()
	r, err := http.NewRequest(req.GetMethod(), req.GetUrl(), bytes.NewReader(req.GetBody()))
	if err != nil {
		log.Fatalf("sdk: cannot create http request from given proto request, err: %v", err)
	}

	for header, values := range req.GetHeader() {
		r.Header[header] = values.Fields
	}

	h.ServeHTTP(w, r)
	// _, _ = io.Copy(os.Stdout, os.Stderr)

	// write response information to sandbox stdout, using for check valid response
	rsp := new(pb.HTTPResponse)
	rsp.Header = make(map[string]*pb.HeaderFields)
	for key, value := range w.header {
		rsp.Header[key] = &pb.HeaderFields{Fields: value}
	}

	rsp.Body = w.buffer.Bytes()
	rsp.Code = int32(w.code)
	bufBytes, err := proto.Marshal(rsp)
	if err != nil {
		log.Fatalf("sdk: cannot handle marshal response")
	}

	written, err := os.Stdout.Write(bufBytes)
	if err != nil {
		log.Fatalf("sdk: cannot handle write response")
	}

	if written != len(bufBytes) {
		log.Fatalf("sdk: given response with length: %d, written: %d", len(bufBytes), written)
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(bufBytes)))

	_, _ = os.Stdout.Write(buf)
}
