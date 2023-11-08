//go:build !debug

package main

import (
	"embed"
)

//go:embed __yostatic
var staticFsYo embed.FS

//go:embed __static
var staticFsApp embed.FS
