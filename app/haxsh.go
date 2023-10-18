package haxsh

var devModeStartMockUsers func()

func Init() {
}

func OnBeforeListenAndServe() {
	if devModeStartMockUsers != nil {
		go devModeStartMockUsers()
	}
}
