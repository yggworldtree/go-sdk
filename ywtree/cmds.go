package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
	"github.com/yggworldtree/go-core/utils"
)

type MessageTopic struct {
	Id     string             `json:"id,omitempty"`
	Path   *bean.TopicPath    `json:"path,omitempty"`
	Sender *bean.CliGroupPath `json:"sender,omitempty"`
	Head   []byte             `json:"head,omitempty"`
	Body   []byte             `json:"body,omitempty"`
	header *utils.Map
}

func (c *MessageTopic) Header() *utils.Map {
	if c.header == nil {
		c.header = utils.NewMaps(c.Head)
	}
	return c.header
}

func (c *Engine) onTopicGet(msg *messages.MessageBox) {
	tph := msg.Info.Args.Get("topicPath")
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
		c.lsr.OnMessage(c, m)
	}
}
func (c *Engine) onNetConnect(msg *messages.MessageBox) {
	code := msg.Info.Args.Get("code")
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
