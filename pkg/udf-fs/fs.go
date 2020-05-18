package udf_fs

/*
#cgo pkg-config: udfread
#include <stdlib.h>

#include <udfread/udfread.h>
*/
import "C"

import (
	"fmt"

	"bazil.org/fuse/fs"
)

type FS struct {
	udf *C.udfread
}

var _ fs.FS = (*FS)(nil)

func NewFS(iso string) (*FS, error) {
	udfread := C.udfread_init()
	if udfread == nil {
		return nil, fmt.Errorf("couldn't initalize udfread\n")
	}

	ret := C.udfread_open(udfread, C.CString(iso))
	if ret != 0 {
		err := fmt.Errorf("failed to open %v\n", iso)
		return nil, err
	}

	return &FS{udf: udfread}, nil
}

func (f *FS) Root() (fs.Node, error) {
	cache, err := makeCache(f.udf, "/")
	if err != nil {
		return nil, err
	}

	d := &dir{
		udf:   f.udf,
		path:  "/",
		cache: cache,
	}

	return d, nil
}

func (f *FS) Release() {
	if f.udf != nil {
		C.udfread_close(f.udf)
	}
}
