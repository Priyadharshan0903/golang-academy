// Package resp reads and writes RESP2, the Redis serialization protocol.
//
// Clients send commands as arrays of bulk strings:
//
//	*2\r\n$3\r\nGET\r\n$4\r\nname\r\n     ->  ["GET", "name"]
//
// Servers reply with one of five types:
//
//	+OK\r\n            simple string
//	-ERR message\r\n   error
//	:42\r\n            integer
//	$5\r\nhello\r\n    bulk string ($-1\r\n is null)
//	*N\r\n...          array of N replies
package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrProtocol is wrapped by every error caused by malformed input.
var ErrProtocol = errors.New("protocol error")

// Limits that protect the server from huge or malicious requests.
const (
	maxArgs     = 1024
	maxBulkSize = 512 * 1024 * 1024 // Redis' own limit for a bulk string
	maxLineSize = 64 * 1024
)

// Reader parses commands from a client connection.
type Reader struct {
	br *bufio.Reader
}

// NewReader wraps r with buffering.
func NewReader(r io.Reader) *Reader { return &Reader{br: bufio.NewReaderSize(r, 16*1024)} }

// ReadCommand reads one command. It accepts the standard array form
// and the "inline" form (plain text like "PING\r\n") that telnet and
// redis-cli in some modes send. It returns io.EOF when the client
// disconnects cleanly between commands.
func (r *Reader) ReadCommand() ([]string, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return r.ReadCommand() // skip empty lines
	}
	if line[0] != '*' {
		return strings.Fields(line), nil // inline command
	}

	n, err := strconv.Atoi(line[1:])
	if err != nil || n < 0 || n > maxArgs {
		return nil, fmt.Errorf("%w: invalid array length %q", ErrProtocol, line)
	}
	args := make([]string, 0, n)
	for range n {
		arg, err := r.readBulk()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func (r *Reader) readBulk() (string, error) {
	line, err := r.readLine()
	if err != nil {
		return "", noEOF(err)
	}
	if len(line) == 0 || line[0] != '$' {
		return "", fmt.Errorf("%w: expected bulk string, got %q", ErrProtocol, line)
	}
	size, err := strconv.Atoi(line[1:])
	if err != nil || size < 0 || size > maxBulkSize {
		return "", fmt.Errorf("%w: invalid bulk length %q", ErrProtocol, line)
	}
	buf := make([]byte, size+2) // payload plus \r\n
	if _, err := io.ReadFull(r.br, buf); err != nil {
		return "", noEOF(err)
	}
	if buf[size] != '\r' || buf[size+1] != '\n' {
		return "", fmt.Errorf("%w: bulk string not terminated by CRLF", ErrProtocol)
	}
	return string(buf[:size]), nil
}

// readLine reads up to \r\n (a bare \n is accepted for inline commands).
func (r *Reader) readLine() (string, error) {
	line, err := r.br.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) || len(line) > maxLineSize {
		return "", fmt.Errorf("%w: line too long", ErrProtocol)
	}
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return "", io.ErrUnexpectedEOF
		}
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(string(line), "\n"), "\r"), nil
}

// noEOF turns io.EOF in the middle of a command into ErrUnexpectedEOF.
func noEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Writer writes replies. Call Flush after each reply (or batch).
type Writer struct {
	bw *bufio.Writer
}

// NewWriter wraps w with buffering.
func NewWriter(w io.Writer) *Writer { return &Writer{bw: bufio.NewWriter(w)} }

// Simple writes a simple string such as OK or PONG.
func (w *Writer) Simple(s string) { w.bw.WriteString("+" + s + "\r\n") }

// Error writes an error reply. By convention it starts with a code such as ERR.
func (w *Writer) Error(msg string) { w.bw.WriteString("-" + msg + "\r\n") }

// Int writes an integer reply.
func (w *Writer) Int(n int64) {
	w.bw.WriteByte(':')
	w.bw.WriteString(strconv.FormatInt(n, 10))
	w.bw.WriteString("\r\n")
}

// Bulk writes a bulk string reply.
func (w *Writer) Bulk(s string) {
	w.bw.WriteByte('$')
	w.bw.WriteString(strconv.Itoa(len(s)))
	w.bw.WriteString("\r\n")
	w.bw.WriteString(s)
	w.bw.WriteString("\r\n")
}

// Null writes the null bulk string, used for missing keys.
func (w *Writer) Null() { w.bw.WriteString("$-1\r\n") }

// Array writes an array of bulk strings, the format commands use.
func (w *Writer) Array(items []string) {
	w.bw.WriteByte('*')
	w.bw.WriteString(strconv.Itoa(len(items)))
	w.bw.WriteString("\r\n")
	for _, it := range items {
		w.Bulk(it)
	}
}

// Flush sends buffered data to the underlying writer.
func (w *Writer) Flush() error { return w.bw.Flush() }
