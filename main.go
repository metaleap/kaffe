package main

import (
	"embed"

	"yo"

	haxsh "haxsh/app"
)

//go:embed __static
var staticFsApp embed.FS

func main() {
	// data, err := staticFsApp.ReadFile("__static/haxsh.html")
	// if err != nil {
	// 	panic(err)
	// }
	// println(">>>>>>>>>" + string(data))
	haxsh.Init(staticFsApp)
	doListenAndServe := yo.Init()
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
