package docker

import (
	"encoding/binary"
	"fmt"
	"io"
)

// idea from https://github.com/ahmetb/dlog

// 01 00 00 00 00 00 00 1f 52 6f 73 65 73 20 61 72 65 ...
// │  ─────┬── ─────┬─────  R  o  s  e  s     a  r  e ...
// │       │        │
// └stdout │        │
//         │        └─ 0x0000001f = 31 bytes (including the \n at the end)
//       unused

const (
	// these should match https://github.com/docker/docker/blob/master/pkg/stdcopy/stdcopy.go
	headerLen          = 8
	SizeByteStartIndex = 4
	SizeByteStopIndex  = SizeByteStartIndex + 4

	initBufLen  = 1024 * 2
	maxLogLen   = 1024 * 64
)

type reader struct {
	r io.Reader // original reader

	// reader state
	begin     bool
	logLen    int
	cursor    int
	buf       []byte
	headerBuf []byte
}

func NewLogReader(r io.Reader) io.Reader {
	return &reader{
		r:         r,
		headerBuf: make([]byte, headerLen),
		buf:       make([]byte, initBufLen),
	}
}

func (r *reader) Read(p []byte) (int, error) {
	if !r.begin {
		if err := r.parse(); err != nil {
			return 0, err
		}
		r.begin = true
	}

	n, err := r.readLog(p) // serve from buf
	if err == io.EOF {
		err = nil
		r.begin = false
	}

	return n, err
}

func (r *reader) readLog(p []byte) (int, error) {
	if r.cursor >= r.logLen {
		return 0, io.EOF
	}

	n := copy(p, r.buf[r.cursor:r.logLen])
	r.cursor += n

	return n, nil
}

func (r *reader) parse() error {
	if n, err := io.ReadFull(r.r, r.headerBuf); err != nil {
		switch err {
		case io.EOF:
			return err // end of the underlying logs stream
		case io.ErrUnexpectedEOF:
			return fmt.Errorf("docker log corrupted prefix. read %d bytes", n)
		default:
			return fmt.Errorf("docker log error reading prefix: %v", err)
		}
	}

	size := int(binary.BigEndian.Uint32(r.headerBuf[SizeByteStartIndex:SizeByteStopIndex]))
	if size > maxLogLen { // safeguard to prevent reading garbage
		return fmt.Errorf("docker log parsed too large: %d (max: %d)", size, maxLogLen)
	}

	// grow buf if necessary
	if size > len(r.buf) {
		r.buf = make([]byte, size)
	}

	// read the log body into buf
	if m, err := io.ReadFull(r.r, r.buf[:size]); err != nil {
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			return fmt.Errorf("docker log corrupt message read %d out of %d bytes: %v", m, size, err)
		default:
			return fmt.Errorf("docker log failed to read message: %v", err)
		}
	}

	r.logLen = size
	r.cursor = 0 // reset cursors for the new message

	return nil
}