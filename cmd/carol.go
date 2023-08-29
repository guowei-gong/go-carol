package main

import (
	"flag"
	"weizicoding.com/carol-dweck"
)

var showLastDoneDaysAgo = flag.Int("d", -1, "显示距离上一次完成练习, 过去了多少天")
var showLevel = flag.String("l", "", "显示练习难度")

func main() {
	flag.Parse()

	_, err := carol.Get()
	if err != nil {
		panic(err)
	}

	// TODO: 解析控制台打印格式
}
