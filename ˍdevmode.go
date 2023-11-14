//go:build debug

package main

import (
	"io/fs"
	"os"
)

var staticFs_Yo fs.FS = os.DirFS(".")
var staticFs_App fs.FS = os.DirFS(".")
