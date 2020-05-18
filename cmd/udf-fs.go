package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fs/fstestutil"

	udf_fs "github.com/sjpotter/udf-fs/pkg/udf-fs"
)

var (
	progName = filepath.Base(os.Args[0])
	debug    = flag.Bool("debug", false, "verbose fuse debugging")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s [-debug] <iso file> <mount point>\n", progName)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	if *debug {
		fstestutil.DebugByDefault()
	}

	file := flag.Arg(0)
	mountpoint := flag.Arg(1)

	if err := mount(file, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(file, mountpoint string) error {
	c, err := fuse.Mount(mountpoint, fuse.AllowOther())
	if err != nil {
		return err
	}
	defer c.Close()

	filesys, err := udf_fs.NewFS(file)
	if err != nil {
		return err
	}
	defer filesys.Release()

	return fs.Serve(c, filesys)
}
