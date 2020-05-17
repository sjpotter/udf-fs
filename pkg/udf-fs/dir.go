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
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&dir{})
var _ = fs.NodeRequestLookuper(&dir{})
var _ = fs.HandleReadDirAller(&dir{})

type direntInfo struct {
	size uint64
	dtType fuse.DirentType
}

type dir struct {
	udf *C.udfread
	path string
	cache map[string]direntInfo
}

func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Size = 0
	attr.Mode = os.ModeDir | 0755
	attr.Valid = 24*7*365*time.Hour

	return nil
}

func (d *dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	di, ok := d.cache[req.Name]
	if !ok {
		return nil, fmt.Errorf("%v does not exist", filepath.Join(d.path, req.Name))
	}

	switch di.dtType {
	case fuse.DT_File:
		return &file{udf: d.udf, path: filepath.Join(d.path, req.Name), size: di.size}, nil
	case fuse.DT_Dir:
		newDir := filepath.Join(d.path, req.Name)
		cache, err := makeCache(d.udf, newDir)
		if err != nil {
			return nil, err
		}
		return &dir{udf: d.udf, path: newDir, cache: cache}, nil
	default:
		err := fmt.Errorf("Unnown directory entry type: %v", di.dtType)
		return nil, err
	}
}

func makeCache(udf *C.udfread, newDir string) (map[string]direntInfo, error) {
	cache := make(map[string]direntInfo)

	dir := C.udfread_opendir(udf, C.CString(newDir))
	if dir == nil {
		err := fmt.Errorf("couldn't opendir: %v", newDir)
		return nil, err
	}
	defer C.udfread_closedir(dir)

	for {
		var dirent = &C.struct_udfread_dirent{}
		dirent = C.udfread_readdir(dir, dirent)
		if dirent == nil {
			break
		}
		var info direntInfo
		switch dirent.d_type {
		case C.UDF_DT_DIR:
			info.dtType = fuse.DT_Dir
		case C.UDF_DT_REG:
			info.dtType = fuse.DT_File
			size, err := fileSize(udf, filepath.Join(newDir, C.GoString(dirent.d_name)))
			if err != nil {
				fmt.Printf("Skipping %v as couldn't determine file size\n", C.GoString(dirent.d_name))
				continue
			}
			info.size = size
		default:
			fmt.Printf("Unknown type: %v\n", dirent.d_type)
			continue
		}

		cache[C.GoString(dirent.d_name)] = info
	}

	return cache, nil
}

func fileSize(udf *C.udfread, file string) (uint64, error) {
	fh := C.udfread_file_open(udf, C.CString(file))
	if fh == nil {
		err := fmt.Errorf("failed to open %v", file)
		return 0, err
	}
	defer C.udfread_file_close(fh)

	size := C.udfread_file_size(fh)
	if size == -1 {
		err := fmt.Errorf("couldn't determine file size for %v", file)
		return 0, err
	}

	return uint64(size), nil
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var ret []fuse.Dirent

	for name, di := range d.cache {
		dirent := fuse.Dirent{
			Name: name,
			Type: di.dtType,
		}

		ret = append(ret, dirent)
	}

	return ret, nil
}