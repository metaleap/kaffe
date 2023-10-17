package main

import (
	"embed"

	"yo"

	haxsh "haxsh/app"
)

//go:embed __static
var staticFsApp embed.FS

//go:embed __yostatic
var staticFsYo embed.FS

func main() {
	haxsh.Init()
	doListenAndServe := yo.Init(staticFsApp, "__static", staticFsYo)
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
