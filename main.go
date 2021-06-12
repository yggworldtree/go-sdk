package main

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-sdk/ywtree"
	"time"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	println("this is test for sdk")
	egn := ywtree.NewEngine(nil, &TmpLsr{}, &ywtree.Config{
		Host: "localhost:7000",
		Org:  "mgr",
		Name: "test",
	})
	egn.RegHbtpFun(1, testFun)
	err := egn.Run()
	if err != nil {
		println("mgr.run err:" + err.Error())
	}
}

type TmpLsr struct{}

func (c *TmpLsr) OnConnect(egn *ywtree.Engine) {
	egn.SubTopic([]*bean.TopicInfo{
		{
			Path:  bean.NewTopicPath("mgr", "cpu_info"),
			Safed: false,
		},
	})
	go func() {
		ls, err := egn.GroupClients("mgr", "test")
		if err != nil {
			logrus.Debugf("GroupClients err:%v", err)
		} else {
			for i, v := range ls {
				logrus.Debugf("GroupClients cli[%d]:%s/%s:%s(%s),", i, v.Org, v.Name, v.Alias, v.Id)
			}
		}
		egn.PushTopic(bean.NewTopicPath("mgr", "cpu_info"), []byte("第一次发送"))
		time.Sleep(time.Second * 3)
		logrus.Debugf("PushTopic!!!!!!!!!!!!!!!")
		egn.PushTopic(bean.NewTopicPath("mgr", "cpu_info"), []byte("第二次发送"))
		logrus.Debugf("HbtpGrpcRequest!!!!!!!!!!!!!!!")
		req, err := egn.HbtpGrpcRequest(bean.NewCliGroupPath("mgr", "test"), 1, "")
		if err != nil {
			logrus.Debugf("HbtpGrpcRequest err:%v", err)
			return
		}
		defer req.Close()
		req.ReqHeader().Set("code", "1234567")
		err = req.Do(nil, nil)
		if err != nil {
			logrus.Debugf("HbtpGrpcRequest do err:%v", err)
			return
		}
		logrus.Debugf("HbtpGrpcRequest res(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}()
}
func (c *TmpLsr) OnDisconnect(egn *ywtree.Engine) {

}
func (c *TmpLsr) OnMessage(egn *ywtree.Engine, pth *bean.TopicPath, data []byte) {
	logrus.Debugf("OnMessage:%s", pth.String())
	switch pth.String() {
	case bean.NewTopicPath("mgr", "cpu_info").String():
		logrus.Debugf("OnMessage data:%s", string(data))
	}
}
func (c *TmpLsr) OnBroadcast(egn *ywtree.Engine) {}

func testFun(c *hbtp.Context) {
	code := c.ReqHeader().GetString("code")
	logrus.Debugf("grpc testFun code:%s", code)
	c.ResString(hbtp.ResStatusOk, "ok")
}
