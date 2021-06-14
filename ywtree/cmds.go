package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
)

func (c *Engine) onTopicGet(msg *messages.MessageBox) *messages.ReplyInfo {
	tph := msg.Info.Args.Get("topicPath")
	if tph == "" {
		logrus.Debugf("onTopicGet param err1")
		return nil
	}
	pth, err := bean.ParseTopicPath(tph)
	//tph := utils.NewMaps(tp)
	//pth := bean.NewTopicPath(tph.GetString("nameSpace"), tph.GetString("key"), tph.GetString("tag"))
	if err != nil {
		logrus.Debugf("onTopicGet param err2:" + err.Error())
		return nil
	}
	m := &MessageTopic{
		Id:   msg.Info.Id,
		Path: pth,
		Head: msg.Head,
		Body: msg.Body,
	}
	if msg.Info.Sender != "" {
		cpth, err := bean.ParseCliGroupPath(msg.Info.Sender)
		if err == nil {
			m.Sender = cpth
		}
	}
	if c.lsr != nil {
		return c.lsr.OnMessage(c, m)
	}
	return nil
}
func (c *Engine) onNetConnect(msg *messages.MessageBox) *messages.ReplyInfo {
	code := msg.Info.Args.Get("code")
	if code == "" {
		logrus.Debugf("onNetConnect param err1")
		return nil
	}
	req := c.newHbtpReq("GrpcClientRes")
	defer req.Close()
	req.ReqHeader().Set("code", code)
	err := req.Do(c.ctx, nil)
	if err != nil {
		logrus.Debugf("onTopicGet param err2:" + err.Error())
		return nil
	}
	if req.ResCode() != hbtp.ResStatusOk {
		logrus.Debugf("onNetConnect server err(%d):%s",
			req.ResCode(), string(req.ResBodyBytes()))
		return nil
	}
	go c.netHandleConn(req.Conn(true))
	return nil
}
