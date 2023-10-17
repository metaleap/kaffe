package haxsh

import (
	"io/fs"
	yosrv "yo/srv"
)

func Init(staticFsApp fs.FS) {
	yosrv.StaticFileServes["__static"] = staticFsApp
}

func OnBeforeListenAndServe() {
}
