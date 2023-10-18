//go:build debug

package haxsh

import "yo/util/str"

const mockFilesDirPath = "__static/mockfiles"

var mockPostFiles = []string{
	"vid1.webm",
	"vid2.mp4",
	"vid3.mp4",
	"post1.jpg",
	"post10.png",
	"post11.jpg",
	"post12.jpg",
	"post13.png",
	"post14.jpg",
	"post15.jpg",
	"post16.png",
	"post17.png",
	"post18.png",
	"post19.jpg",
	"post2.jpg",
	"post20.png",
	"post21.webp",
	"post22.jpg",
	"post23.png",
	"post24.jpg",
	"post25.jpg",
	"post26.png",
	"post27.jpeg",
	"post28.jpg",
	"post29.jpg",
	"post3.jpg",
	"post30.jpg",
	"post31.webp",
	"post4.jpg",
	"post5.jpg",
	"post6.jpg",
	"post7.jpg",
	"post8.jpg",
	"post9.jpg",
}

func init() {
	devModeStartMockUsers = func() {
		for nick_name, pic_file_name := range (str.Dict{
			"foo1@bar.baz": "user1.jpg",
			"foo2@bar.baz": "user2.png",
			"foo3@bar.baz": "user3.jpg",
			"foo4@bar.baz": "user4.png",
			"foo5@bar.baz": "user5.jpg",
			"foo6@bar.baz": "user6.png",
			"foo7@bar.baz": "user7.jpg",
			"foo8@bar.baz": "user8.png",
		}) {
			_, _ = nick_name, pic_file_name
		}
	}
}

func startMockUser(i int) {
}
