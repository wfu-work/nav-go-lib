package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"

	"github.com/wfu-work/nav-go-lib/libs"
)

//export mergeNav
func mergeNav(argv *C.char) *C.char {
	args := C.GoString(argv)
	outPath := libs.Nav(args)
	return C.CString(outPath)
}

//export FreeC
func FreeC(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func main() {}
