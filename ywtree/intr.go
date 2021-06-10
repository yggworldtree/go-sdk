package ywtree

import "github.com/yggworldtree/go-core/bean"

type IYWTListener interface {
	OnConnect(c *Engine)
	OnDisconnect(c *Engine)
	OnMessage(c *Engine, pth *bean.TopicPath, data []byte)
	OnBroadcast(c *Engine)
}
