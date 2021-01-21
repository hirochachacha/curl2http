package curl2http

import (
	"bufio"
	"io"
	"strings"
)

type asciiReader struct {
	sc   *bufio.Scanner
	text string
}

func newAsciiReader(r io.Reader) io.Reader {
	return &asciiReader{
		sc: bufio.NewScanner(r),
	}
}

func (r *asciiReader) Read(p []byte) (int, error) {
	if r.text == "" {
		if !r.sc.Scan() {
			return -1, io.EOF
		}
		if err := r.sc.Err(); err != nil {
			return -1, err
		}
		r.text = r.sc.Text()
	}
	i := copy(p, r.text)
	r.text = r.text[i:]
	return i, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
