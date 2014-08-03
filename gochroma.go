// Package gochroma provides a high-level API to the acoustic
// fingerprinting library chromaprint.
package gochroma

import (
	"github.com/go-fingerprint/fingerprint"
	"github.com/go-fingerprint/gochroma/chromaprint"
	"io"
)

const (
	seconds       = 10
	minmaxseconds = 120
)

// Available algorithms for fingerprinting.
const (
	Algorithm1 = chromaprint.CHROMAPRINT_ALGORITHM_TEST1
	Algorithm2 = chromaprint.CHROMAPRINT_ALGORITHM_TEST2
	Algorithm3 = chromaprint.CHROMAPRINT_ALGORITHM_TEST3
	Algorithm4 = chromaprint.CHROMAPRINT_ALGORITHM_TEST4

	AlgorithmDefault = chromaprint.CHROMAPRINT_ALGORITHM_DEFAULT
)

// A Printer is a fingerprint.Calculator backed by libchromaprint.
type Printer struct {
	context *chromaprint.ChromaprintContext
}

// New creates new Printer. Returned Printer must be closed
// with Close(). Note that if libchromaprint is compiled with FFTW,
// New should be called only from one goroutine at a time.
func New(algorithm int) (p *Printer) {
	return &Printer{chromaprint.NewChromaprint(algorithm)}
}

// Close existing Printer.
func (p *Printer) Close() {
	p.context.Free()
}

// Fingerprint implements fingerprint.Calculator interface.
func (p *Printer) Fingerprint(i fingerprint.RawInfo) (fprint string, err error) {
	if err = p.prepare(i); err != nil {
		return
	}
	fprint, err = p.context.GetFingerprint()
	return
}

// RawFingerprint implements fingerprint.Calculator interface.
func (p *Printer) RawFingerprint(i fingerprint.RawInfo) (fprint []int32, err error) {
	if err = p.prepare(i); err != nil {
		return
	}
	fprint, err = p.context.GetRawFingerprint()
	return
}

func (p *Printer) prepare(i fingerprint.RawInfo) error {
	if i.MaxSeconds < minmaxseconds {
		i.MaxSeconds = minmaxseconds
	}
	ctx := p.context
	rate, channels := i.Rate, i.Channels
	if err := ctx.Start(int(rate), int(channels)); err != nil {
		return err
	}
	numbytes := 2 * seconds * rate * channels
	buf := make([]byte, numbytes)
	for total := uint(0); total <= i.MaxSeconds; total += seconds {
		read, err := i.Src.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if read == 0 {
			break
		}
		if err := ctx.Feed(buf[:read]); err != nil {
			return err
		}
	}
	if err := ctx.Finish(); err != nil {
		return err
	}
	return nil
}
