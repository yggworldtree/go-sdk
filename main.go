package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-sdk/ywtree"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	println("this is test for sdk")
	egn := ywtree.NewEngine(nil, &ywtree.Config{
		Host: "localhost:7000",
		Org:  "mgr",
		Name: "test",
	})
	err := egn.Run()
	if err != nil {
		println("mgr.run err:" + err.Error())
	}
}
