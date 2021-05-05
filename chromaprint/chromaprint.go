// Package chromaprint is a Go binding to
// acoustic fingerprinting library chromaprint.
package chromaprint

// #cgo pkg-config: fftw3
// #cgo LDFLAGS: -lchromaprint -lm -lstdc++
// #include <stdlib.h>
// #include <chromaprint.h>
//
// void **allocate_mem() {
//		return malloc(sizeof(void*));
// }
import "C"

import (
	"errors"
	"reflect"
	"time"
	"unsafe"
)

// Possible errors that may occur during function calls.
var (
	ErrSetOption = errors.New("chromaprint: failed to set option")
	ErrStart     = errors.New("chromaprint: failed to restart chromaprint")
	ErrFeed      = errors.New("chromaprint: failed to send data to the fingerprint calculator")
	ErrFinish    = errors.New("chromaprint: failed to process buffered audiodata")
	ErrFprint    = errors.New("chromaprint: failed to calculate compressed fingerprint")
	ErrRawFprint = errors.New("chromaprint: failed to calculate raw fingerprint")
)

// Available algorithms for fingerprinting.
const (
	CHROMAPRINT_ALGORITHM_TEST1 = C.CHROMAPRINT_ALGORITHM_TEST1
	CHROMAPRINT_ALGORITHM_TEST2 = C.CHROMAPRINT_ALGORITHM_TEST2
	CHROMAPRINT_ALGORITHM_TEST3 = C.CHROMAPRINT_ALGORITHM_TEST3
	CHROMAPRINT_ALGORITHM_TEST4 = C.CHROMAPRINT_ALGORITHM_TEST4

	CHROMAPRINT_ALGORITHM_DEFAULT = C.CHROMAPRINT_ALGORITHM_DEFAULT
)

// ChromaprintVersion returns version numper of chromaprint library.
func ChromaprintVersion() string {
	return C.GoString(C.chromaprint_get_version())
}

// A ChromaprintContext holds all data structures required for fingerprint
// calculation.
type ChromaprintContext struct {
	context   *C.ChromaprintContext
	algorithm int
}

// NewChromaprint creates new CromaprintContext with given algorithm.
// Returned ChromaprintContext must be freed with Free().
func NewChromaprint(algorithm int) *ChromaprintContext {
	return &ChromaprintContext{C.chromaprint_new(C.int(algorithm)), algorithm}
}

// Free ChromaprintContext
func (ctx *ChromaprintContext) Free() {
	C.chromaprint_free(ctx.context)
}

// Algorithm returns selected algorithm, one of the
// CHROMAPRINT_ALGORITHM_TEST1 etc.
func (ctx *ChromaprintContext) Algorithm() int {
	return ctx.algorithm
}

// SetOption sets a configuration option for the selected algorithm.
// Currently only one option is available: 'silence_threshold'
// with possible values 0-32767.
func (ctx *ChromaprintContext) SetOption(name string, value int) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if int(C.chromaprint_set_option(ctx.context, cname, C.int(value))) < 1 {
		return ErrSetOption
	}
	return nil
}

// Start restarts the computation of a fingerprint with a new audio stream.
func (ctx *ChromaprintContext) Start(sampleRate, numChannels int) error {
	if int(C.chromaprint_start(ctx.context, C.int(sampleRate), C.int(numChannels))) < 1 {
		return ErrStart
	}
	return nil
}

// Feed sends audio data, encoded in 16-bit signed integers with
// native byte-order, which in most architectures such as Intel x86
// and x86-64 is little-endian.
func (ctx *ChromaprintContext) Feed(data []byte) error {
	if int(C.chromaprint_feed(ctx.context, (*C.int16_t)(unsafe.Pointer(&data[0])), C.int(len(data)/2))) < 1 {
		return ErrFeed
	}
	return nil
}

// Finish processes any remaining buffered data and calculates the fingerprint.
func (ctx *ChromaprintContext) Finish() error {
	if r := C.chromaprint_finish(ctx.context); int(r) < 1 {
		return ErrFinish
	}
	return nil
}

// GetFingerprint returns the calculated fingerprint as a compressed string.
func (ctx *ChromaprintContext) GetFingerprint() (fprint string, err error) {
	ptr := (**C.char)(unsafe.Pointer(C.allocate_mem()))
	defer C.free(unsafe.Pointer(ptr))

	if int(C.chromaprint_get_fingerprint(ctx.context, ptr)) < 1 {
		return "", ErrFprint
	}

	fprint = C.GoString(*ptr)
	C.chromaprint_dealloc(unsafe.Pointer(*ptr))
	return
}

// GetRawFingerprint returns the calculated fingerprint
// as a slice of 32-bit integers.
func (ctx *ChromaprintContext) GetRawFingerprint() (fprint []int32, err error) {
	ptr := C.allocate_mem()
	defer C.free(unsafe.Pointer(ptr))

	var size C.int
	if int(C.chromaprint_get_raw_fingerprint(ctx.context, (**C.uint32_t)(unsafe.Pointer(ptr)), &size)) < 1 {
		return nil, ErrRawFprint
	}

	fprint = goInt32s(*ptr, int(size))
	return
}

// GetItemDuration returns the duration of one item in the raw fingerprint in milliseconds.
func (ctx *ChromaprintContext) GetItemDuration() time.Duration {
	ms := int(C.chromaprint_get_item_duration_ms(ctx.context))
	return time.Duration(ms) * time.Millisecond
}

// GetItemDurationSamples returns the duration of one item in the raw fingerprint in samples.
func (ctx *ChromaprintContext) GetItemDurationSamples() int {
	return int(C.chromaprint_get_item_duration(ctx.context))
}

func goInt32s(ptr unsafe.Pointer, size int) (ints []int32) {
	hdr := reflect.SliceHeader{
		Data: uintptr(ptr),
		Len:  size,
		Cap:  size,
	}
	cints := *(*[]int32)(unsafe.Pointer(&hdr))
	ints = make([]int32, size)
	copy(ints, cints)
	return
}
