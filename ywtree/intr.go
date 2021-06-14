package ywtree

import "github.com/yggworldtree/go-core/messages"

type IYWTListener interface {
	OnConnect(c *Engine)
	OnDisconnect(c *Engine)
	OnMessage(c *Engine, msg *MessageTopic) *messages.ReplyInfo
	OnBroadcast(c *Engine, msg *messages.MessageBox) *messages.ReplyInfo
}
