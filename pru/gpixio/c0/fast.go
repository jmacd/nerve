//go:build !darwin

package gpixio

// #cgo CFLAGS: -mfloat-abi=hard
// #include "fast.h"
import "C"

func Copy2() {
	C.CopyPixels(nil, nil)
}
