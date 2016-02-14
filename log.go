package main

import (
	"fmt"
	"os"
)

var infoPrefix = []byte("[Info] ")
var warnPrefix = []byte("\x1b[33m[Warning] ")
var errorPrefix = []byte("\x1b[31m[Error] ")
var fatalPrefix = []byte("\x1b[35m[Fatal] ")
var newline = []byte("\x1b[0m\n")

// Infof formats and prints the given message into stdout
func Infof(message string, args ...interface{}) {
	os.Stdout.Write(infoPrefix)
	os.Stdout.Write([]byte(fmt.Sprintf(message, args...)))
	os.Stdout.Write(newline)
}

// Warnf formats and prints the given message into stdout with a yellow color
func Warnf(message string, args ...interface{}) {
	os.Stdout.Write(warnPrefix)
	os.Stdout.Write([]byte(fmt.Sprintf(message, args...)))
	os.Stdout.Write(newline)
}

// Errorf formats and prints the given message into stderr with a red color
func Errorf(message string, args ...interface{}) {
	os.Stdout.Write(errorPrefix)
	os.Stdout.Write([]byte(fmt.Sprintf(message, args...)))
	os.Stdout.Write(newline)
}

// Fatalf formats and prints the given message into stderr with a purple color
func Fatalf(message string, args ...interface{}) {
	os.Stdout.Write(fatalPrefix)
	os.Stdout.Write([]byte(fmt.Sprintf(message, args...)))
	os.Stdout.Write(newline)
}
