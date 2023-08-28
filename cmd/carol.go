package main

import "flag"

var showLastDoneDaysAgo = flag.Int("d", -1, "显示距离上一次完成练习, 过去了多少天")
var showLevel = flag.String("l", "", "显示练习难度")

func main() {
	flag.Parse()

}
