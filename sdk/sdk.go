package sdk

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/hnimtadd/run/pb/v1"
)

// var (
// 	requestBuffer  []byte
// 	responseBuffer []byte
// )

// type request struct {
// 	Header http.Header
// 	Method string
// 	URL    string
// 	Body   []byte
// }

type responseWriter struct {
	buffer *bytes.Buffer
	code   int
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		buffer: new(bytes.Buffer),
	}
}

func (w *responseWriter) Header() http.Header {
	return http.Header{}
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	return w.buffer.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.code = status
}

// unmarshal the mashaled request and process
func Handle(h http.Handler) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("sdk: cannot read stdin, err: %v", err)
	}

	req := new(pb.HTTPRequest)
	if err := proto.Unmarshal(b, req); err != nil {
		log.Fatalf("sdk: cannot unmarshal the proto request, err: %v", err)
	}

	r, err := http.NewRequest(req.GetMethod(), req.GetUrl(), bytes.NewReader(req.GetBody()))
	if err != nil {
		log.Fatalf("sdk: cannot create http request from given proto request, err: %v", err)
	}

	for header, values := range req.GetHeader() {
		r.Header[header] = values.Fields
	}

	w := newResponseWriter()

	h.ServeHTTP(w, r)
	// write response to sandbox stdout
	if _, err := io.Copy(os.Stdout, w.buffer); err != nil {
		log.Panicf("sdk, cannot write response to sandbox stdout, err: %v", err)
	}

	// write response information to sandbox stdout, using for check valid response
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(w.code))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(w.buffer.Len()))
	os.Stdout.Write(buf)
}
