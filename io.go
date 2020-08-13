package admin

import (
	"io"
)

type TrimLeftWriter struct {
	W          io.Writer
	Before     func()
	hasContent bool
}

func NewTrimLeftWriter(w io.Writer) *TrimLeftWriter {
	return &TrimLeftWriter{W: w}
}

func (this *TrimLeftWriter) Empty() bool {
	return !this.hasContent
}

func (this *TrimLeftWriter) Write(p []byte) (n int, err error) {
	if this.hasContent {
		return this.W.Write(p)
	}
	var (
		i int
		r byte
	)

	for i, r = range p {
		if r > 32 {
			this.hasContent = true
			if this.Before != nil {
				this.Before()
			}
			return this.W.Write(p[i:])
		}
	}
	return
}

func (this *TrimLeftWriter) WriteString(p string) (n int, err error) {
	return this.Write([]byte(p))
}
