package main

import (
	"yo"

	kaffe "kaffe/app"
)

func main() {
	kaffe.Init() // keep in `main()`, dont move to `init()`
	doListenAndServe := yo.Init(staticFsYo, staticFsApp)
	kaffe.OnBeforeListenAndServe()
	doListenAndServe()
}
