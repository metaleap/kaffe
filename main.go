package main

import (
	"embed"

	"yo"

	haxsh "haxsh/app"
)

//go:embed __static
var staticFsApp embed.FS

func main() {
	haxsh.Init(staticFsApp)
	doListenAndServe := yo.Init()
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
