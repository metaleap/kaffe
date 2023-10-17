package main

import (
	"yo"
)

func main() {
	run := yo.Init()
	{
		// in here goes anything to occur between above Init and below never-returning listen-and-serve call
	}
	run()
}
