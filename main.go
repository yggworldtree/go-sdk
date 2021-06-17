package main

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
	"github.com/yggworldtree/go-sdk/ywtree"
	"time"
)

var egn *ywtree.Engine

func main() {
	hbtp.Debug = true
	println("this is test for sdk")
	egn = ywtree.NewEngine(nil, &TmpLsr{}, &ywtree.Config{
		Host: "localhost:7000",
		Org:  "mgr",
		Name: "test",
		Sign: "123456",
	})
	egn.RegHbtpFun(1, testFun)
	err := egn.Run()
	if err != nil {
		println("mgr.run err:" + err.Error())
	}
}

type TmpLsr struct{}

func (c *TmpLsr) OnConnect(egn *ywtree.Engine) {
	pthCpu := bean.NewTopicPath("mgr/test", "cpu_info")
	egn.SubTopic([]*bean.TopicInfo{
		{
			Path:  pthCpu.String(),
			Safed: false,
		},
	})
	go func() {
		ls, err := egn.GroupClients("mgr", "test")
		if err != nil {
			hbtp.Debugf("GroupClients err:%v", err)
		} else {
			for i, v := range ls {
				hbtp.Debugf("GroupClients cli[%d]:%s/%s:%s(%s),", i, v.Org, v.Name, v.Alias, v.Id)
			}
		}
		//egn.CreateBucket("test")
		//egn.SetBucketParam("test", "sends", []byte("第一次发送s"))
		egn.PushTopic(pthCpu, []byte("第一次发送"))
		time.Sleep(time.Second * 3)
		bts, err := egn.GetBucketParam("test", "sends")
		if err == nil {
			hbtp.Debugf("bucket get test/sends:%s", string(bts))
		}
		//egn.UnSubTopic(pthCpu)
		hbtp.Debugf("PushTopic!!!!!!!!!!!!!!!")
		egn.PushTopic(pthCpu, []byte("第二次发送"))
		hbtp.Debugf("HbtpGrpcRequest!!!!!!!!!!!!!!!")
		req, err := egn.HbtpGrpcRequest(bean.NewCliGroupPath("mgr", "test"), 1, "")
		if err != nil {
			hbtp.Debugf("HbtpGrpcRequest err:%v", err)
			return
		}
		defer req.Close()
		req.ReqHeader().Set("code", "1234567")
		err = req.Do(nil, nil)
		if err != nil {
			hbtp.Debugf("HbtpGrpcRequest do err:%v", err)
			return
		}
		hbtp.Debugf("HbtpGrpcRequest res(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}()
}
func (c *TmpLsr) OnDisconnect(egn *ywtree.Engine) {

}
func (c *TmpLsr) OnMessage(egn *ywtree.Engine, msg *ywtree.MessageTopic) *messages.ReplyInfo {
	pths := msg.Path.String()
	hbtp.Debugf("OnMessage:%s,from:%s", pths, msg.Sender.String())
	switch pths {
	case bean.NewTopicPath("mgr/test", "cpu_info").String():
		hbtp.Debugf("OnMessage data:%s", string(msg.Body))
	}

	return nil
}
func (c *TmpLsr) OnBroadcast(egn *ywtree.Engine, msg *messages.MessageBox) *messages.ReplyInfo {
	return nil
}

func testFun(c *hbtp.Context) {
	code := c.ReqHeader().GetString("code")
	hbtp.Debugf("grpc testFun code:%s", code)
	c.ResString(hbtp.ResStatusOk, egn.Alias()+":ok")
}
