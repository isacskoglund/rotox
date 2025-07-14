package telemetry

import (
	"io"
	"time"

	"github.com/isacskoglund/rotox/internal/common"
)

const (
	batchSize = 32 * 1024
)

type emitFunc func(start time.Time, stop time.Time, n uint64)

func NewTelemetryConn(next common.Conn, emit emitFunc) common.Conn {
	base := connBase{next: next}
	writer := emitWriter{
		next: next,
		emit: emit,
	}
	wt, isWriterTo := next.(io.WriterTo)
	rf, isReaderFrom := next.(io.ReaderFrom)

	type writerTo io.WriterTo
	if isWriterTo && isReaderFrom {
		return &struct {
			connBase
			emitWriter
			writerTo
			emitReaderFrom
		}{
			connBase:   base,
			emitWriter: writer,
			writerTo:   wt,
			emitReaderFrom: emitReaderFrom{
				next: rf,
				emit: emit,
			},
		}
	}

	if isWriterTo {
		return &struct {
			connBase
			emitWriter
			writerTo
		}{
			connBase:   base,
			emitWriter: writer,
			writerTo:   wt,
		}
	}

	if isReaderFrom {
		return &struct {
			connBase
			emitWriter
			emitReaderFrom
		}{
			connBase:   base,
			emitWriter: writer,
			emitReaderFrom: emitReaderFrom{
				next: rf,
				emit: emit,
			},
		}
	}

	return &struct {
		connBase
		emitWriter
	}{
		connBase:   base,
		emitWriter: writer,
	}
}

type connBase struct {
	next common.Conn
}

func (base *connBase) Read(dst []byte) (int, error) {
	return base.next.Read(dst)
}

func (base *connBase) Close() error {
	return base.next.Close()
}

func (base *connBase) Name() string {
	return base.next.Name()
}

type emitReader struct {
	next io.Reader
	emit emitFunc
}

func (r *emitReader) Read(dst []byte) (int, error) {
	m := 0
	for i := 0; i < len(dst); i += batchSize {
		start := time.Now()
		n, err := r.next.Read(chunk(dst, i, i+batchSize))
		stop := time.Now()
		r.emit(start, stop, uint64(n))
		m += n
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

type emitWriter struct {
	next io.Writer
	emit emitFunc
}

func (w *emitWriter) Write(dst []byte) (int, error) {
	m := 0
	for i := 0; i < len(dst); i += batchSize {
		start := time.Now()
		n, err := w.next.Write(chunk(dst, i, i+batchSize))
		stop := time.Now()
		w.emit(start, stop, uint64(n))
		m += n
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

type emitReaderFrom struct {
	next io.ReaderFrom
	emit emitFunc
}

func (rf *emitReaderFrom) ReadFrom(r io.Reader) (int64, error) {
	r = &emitReader{
		next: r,
		emit: rf.emit,
	}
	n, err := rf.next.ReadFrom(r)
	return n, err
}

func chunk[T any](items []T, start int, stop int) []T {
	begin := max(0, start)
	end := min(len(items), stop)
	if begin >= end {
		return nil
	}
	return items[begin:end]
}
