package ywtree

import "github.com/yggworldtree/go-core/messages"

type IYWTListener interface {
	OnConnect(egn *Engine)
	OnDisconnect(egn *Engine)
	OnMessage(egn *Engine, msg *MessageTopic) *messages.ReplyInfo
	OnBroadcast(egn *Engine, msg *messages.MessageBox) *messages.ReplyInfo
}
