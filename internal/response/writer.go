package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type WriterState int

const (
	writerStateRequestLine   WriterState = 1
	writerStateHeaders       WriterState = 2
	writerStateBody          WriterState = 3
)

type Writer struct {
	writerState WriterState
	writer      io.Writer
}

func NewWriter(writerToWrap io.Writer) *Writer {
	return &Writer {
		writerState: writerStateRequestLine,
		writer:      writerToWrap,
	}
}

func (w *Writer) WriteRequestLine(statusCode StatusCode) error {
	if w.writerState != writerStateRequestLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()

	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()

	for k, v := range h {
		_, err := w.writer.Write(fmt.Appendf(nil, "%s: %s\r\n", k, v))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}

	return w.writer.Write(p)
}
