package main

import (
	"flag"
	"weizicoding.com/carol"
)

var showLevel = flag.String("l", "", "显示练习难度, easy; medium; hard")
var sortByColumn = flag.Int("s", 2, "排序依据, 1:名称; 2:上一次完成时间; 3:完成次数")
var showLastDoneDaysAgo = flag.Int("d", -1, "显示距离上一次完成练习, 过去了多少天")

func main() {
	flag.Parse()

	practices, err := carol.Get()
	if err != nil {
		panic(err)
	}

	// 解析控制台打印格式
	carol.Print(practices, *showLastDoneDaysAgo, *sortByColumn, *showLevel)
}
