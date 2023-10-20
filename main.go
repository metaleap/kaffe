package main

import (
	"yo"

	haxsh "haxsh/app"
)

func main() {
	haxsh.Init() // keep in `main()`, dont move to `init()`
	doListenAndServe := yo.Init(staticFsApp, staticFsYo)
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
