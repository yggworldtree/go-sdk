package ywtree

import (
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean/clientBean"
)

type mpCliFun func(cli *Engine, msg *clientBean.MessageBox)

var mpCliFn = map[string]mpCliFun{
	"msg/heart": onHeart,
}

func onHeart(cli *Engine, msg *clientBean.MessageBox) {
	logrus.Debugf("get msg heart:%s", msg.Head.Id)
}
