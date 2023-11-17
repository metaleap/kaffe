package kaffe

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/url"

	. "yo/ctx"
	"yo/util/sl"
)

// assumes a square image, as we do square avatar uploads. only for favicons, since we css-round otherwise.
func imageRoundedSvgOfSquareImage(_ *Ctx, src_raw []byte) (fileExt string, fileSrc []byte) {
	_, format, _ := image.Decode(bytes.NewReader(src_raw))

	src_b64 := base64.StdEncoding.EncodeToString(src_raw)
	data_url := "data:image/" + format + ";base64," + url.PathEscape(src_b64) // would love to do without and have direct href, but cors & co...
	svg_raw := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100" height="100" viewBox="0 0 100 100">
					<clipPath id="c"><circle cx="50" cy="50" r="50" /></clipPath>
					<image x="0" y="0" width="100" height="100" clip-path="url(#c)" href="` + data_url + `"/>
				</svg>`
	return ".svg", []byte(svg_raw)
}

func imageSquared(srcRaw []byte) []byte {
	known_formats := []string{"png", "jpeg", "gif"}
	img, format, _ := image.Decode(bytes.NewReader(srcRaw))
	if img_old, _ := img.(interface {
		image.Image
		SubImage(r image.Rectangle) image.Image
	}); (img_old != nil) && sl.Has(known_formats, format) {
		if sub_rect := img_old.Bounds(); sub_rect.Dx() != sub_rect.Dy() {
			if w, h := sub_rect.Dx(), sub_rect.Dy(); w > h {
				sub_rect = image.Rect((w-h)/2, 0, ((w-h)/2)+h, h)
			} else { // h > w
				sub_rect = image.Rect(0, (h-w)/2, w, ((h-w)/2)+w)
			}
			img = img_old.SubImage(sub_rect)
			var buf bytes.Buffer
			switch format {
			case "png":
				if err := png.Encode(&buf, img); err != nil {
					buf.Reset()
				}
			case "jpeg":
				if err := jpeg.Encode(&buf, img, nil); err != nil {
					buf.Reset()
				}
			case "gif":
				if err := gif.Encode(&buf, img, nil); err != nil {
					buf.Reset()
				}
			}
			if buf.Len() > 0 {
				srcRaw = buf.Bytes()
			}
		}
	}
	return srcRaw
}
