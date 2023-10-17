package main

import (
	"yo"

	haxsh "haxsh/app"
)

func main() {
	haxsh.Init()
	doListenAndServe := yo.Init()
	haxsh.OnBeforeListenAndServe()
	doListenAndServe()
}
