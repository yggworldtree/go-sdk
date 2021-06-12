package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
)

func (c *Engine) onTopicGet(msg *messages.MessageBox) {
	tph := msg.Head.Args.GetString("topicPath")
	if tph == "" {
		logrus.Debugf("onTopicGet param err1")
		return
	}
	pth, err := bean.ParseTopicPath(tph)
	//tph := utils.NewMaps(tp)
	//pth := bean.NewTopicPath(tph.GetString("nameSpace"), tph.GetString("key"), tph.GetString("tag"))
	if err != nil {
		logrus.Debugf("onTopicGet param err2:" + err.Error())
		return
	}
	if c.lsr != nil {
		c.lsr.OnMessage(c, pth, msg.Body)
	}
}
func (c *Engine) onNetConnect(msg *messages.MessageBox) {
	code := msg.Head.Args.GetString("code")
	if code == "" {
		logrus.Debugf("onNetConnect param err1")
		return
	}
	req := c.newHbtpReq("GrpcClientRes")
	defer req.Close()
	req.ReqHeader().Set("code", code)
	err := req.Do(c.ctx, nil)
	if err != nil {
		logrus.Debugf("onTopicGet param err2:" + err.Error())
		return
	}
	if req.ResCode() != hbtp.ResStatusOk {
		logrus.Debugf("onNetConnect server err(%d):%s",
			req.ResCode(), string(req.ResBodyBytes()))
		return
	}
	go c.netHandleConn(req.Conn(true))
}
