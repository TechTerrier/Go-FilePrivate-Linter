package package1

import "log"

// fileprivate
var a = "Hello World!"

// fileprivate
var b = "Hello World!" // fileprivate

var c = "Non violation"

type sampleStruct struct {
	Str string
}

// fileprivate
var d = sampleStruct{Str: "test"}

// fileprivate
var e = sampleStruct{Str: "test2"}

func printMessages() {
	log.Println(a)
	log.Println(b)
}

func (s sampleStruct) Print() string {
	return s.Str
}
