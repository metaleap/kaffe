//go:build debug

package main

import (
	"io/fs"
	"os"
)

var staticFsYo fs.FS = os.DirFS(".")
var staticFsApp fs.FS = os.DirFS(".")
