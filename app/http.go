package haxsh

import (
	yosrv "yo/srv"
	. "yo/util"
)

func init() {
	yosrv.AppApiUrlPrefix = "_/"
	yosrv.AppSideStaticRedirectPathFor = func(requestPath string) string {
		is_home := (requestPath == "") // `requestPath` never has a leading slash
		return If(is_home, "__static/home.html", "__static/haxsh.html")
	}
}
