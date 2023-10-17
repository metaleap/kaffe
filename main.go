package main

import (
	"yo"

	haxsh "haxsh/app"
)

func main() {
	haxsh.Init()
	doListenAndServe := yo.Init(staticFsApp, "__static", staticFsYo)
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
