package udf_fs

/*
#cgo pkg-config: udfread
#include <stdlib.h>

#include <udfread/udfread.h>
*/
import "C"
import (
	"context"
	"fmt"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&file{})
var _ = fs.NodeOpener(&file{})

type file struct {
	udf *C.udfread
	path string
	size uint64
}

func (f *file) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Size = f.size
	attr.Mode = 0644
	attr.Valid = 24*7*365*time.Hour

	return nil
}

func (f *file) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	fh := C.udfread_file_open(f.udf, C.CString(f.path))
	if fh == nil {
		return nil, fmt.Errorf("failed to open: %v", f.path)
	}

	return &fileHandle{fh: fh}, nil
}

var _ = fs.Handle(&fileHandle{})
var _ = fs.HandleReleaser(&fileHandle{})
var _ = fs.HandleReader(&fileHandle{})

type fileHandle struct {
	fh *C.UDFFILE
	offset int64
}

func (fh *fileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	C.udfread_file_close(fh.fh)

	return nil
}

func (fh *fileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if req.Offset != fh.offset {
		offset := C.udfread_file_seek(fh.fh, C.long(req.Offset), C.UDF_SEEK_SET)
		if fh.offset < 0 {
			return fmt.Errorf("failed to seek to %v", req.Offset)
		}
		fh.offset = int64(offset)
	}

	buf := C.malloc(C.ulong(req.Size))
	defer C.free(buf)

	n := C.udfread_file_read(fh.fh, buf, C.ulong(req.Size))
	if n < 0 {
		return fmt.Errorf("failed to read file")
	}

	fh.offset += int64(n)

	resp.Data = C.GoBytes(buf, C.int(n))

	return nil
}