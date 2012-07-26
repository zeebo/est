package main

import "fmt"

const (
	timeFormat    = "2006-01-02 15:04:05.999999999 -0700 MST"
	timeFormatLen = len(timeFormat)
)

var whenTemplateString = fmt.Sprintf(`{{.When.Local.Format "%s"}}`, timeFormat)
var whenFormatString = fmt.Sprintf("%% -%ds", timeFormatLen)
