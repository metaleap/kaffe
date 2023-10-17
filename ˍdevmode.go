//go:build debug

package main

import (
	"io/fs"
	"os"
)

var staticFsApp fs.FS = os.DirFS(".")
var staticFsYo fs.FS = os.DirFS(".")
