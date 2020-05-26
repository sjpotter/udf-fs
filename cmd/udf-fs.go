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

	"github.com/sevlyar/go-daemon"

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

	file := flag.Arg(0)
	mountpoint := flag.Arg(1)

	if !*debug {
		cntxt := &daemon.Context{}
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal("Unable to run: ", err)
		}
		if d != nil {
			return
		}
		defer cntxt.Release()
	} else {
		fstestutil.DebugByDefault()
	}

	if err := mount(file, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(file, mountpoint string) error {
	filesys, err := udf_fs.NewFS(file)
	if err != nil {
		return err
	}
	defer filesys.Release()

	c, err := fuse.Mount(mountpoint, fuse.AllowOther())
	if err != nil {
		return err
	}

	defer c.Close()

	return fs.Serve(c, filesys)
}
