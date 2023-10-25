package haxsh

import (
	"encoding/base64"
	"net/url"
)

func imageRoundedSvgOfPng(png_raw []byte) (string, []byte) {
	png_b64 := base64.StdEncoding.EncodeToString(png_raw)
	data_url := "data:image/png;base64," + url.PathEscape(png_b64)
	svg_raw := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100" height="100" viewBox="0 0 100 100">
					<clipPath id="c"><circle cx="50" cy="50" r="50" /></clipPath>
					<image x="0" y="0" width="100" height="100" clip-path="url(#c)" href="` + data_url + `"/>
				</svg>`
	return ".svg", []byte(svg_raw)
}
