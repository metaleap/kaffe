//go:build !debug

package main

import (
	"embed"
)

//go:embed __yostatic
var staticFs_Yo embed.FS

//go:embed __static
var staticFs_App embed.FS
