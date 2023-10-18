package haxsh

var devModeInitMockUsers func()

func Init() {
}

func OnBeforeListenAndServe() {
	if devModeInitMockUsers != nil {
		go devModeInitMockUsers()
	}
}
