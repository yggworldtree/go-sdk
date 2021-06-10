package ywtree

import (
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
	"github.com/yggworldtree/go-core/utils"
)

type mpCliFun func(cli *Engine, msg *messages.MessageBox)

var mpCliFn = map[string]mpCliFun{
	messages.MsgCmdTopicGet: onTopicGet,
}

func onTopicGet(cli *Engine, msg *messages.MessageBox) {
	tp, ok := msg.Head.Args["topicPath"]
	if !ok {
		logrus.Debugf("onTopicGet param err1")
		return
	}
	tph := utils.NewMaps(tp)
	pth := bean.NewTopicPath(tph.GetString("nameSpace"), tph.GetString("key"), tph.GetString("tag"))
	if pth.NameSpace == "" || pth.Key == "" {
		logrus.Debugf("onTopicGet param err2")
		return
	}
	if cli.lsr != nil {
		cli.lsr.OnMessage(cli, pth, msg.Body)
	}
}
