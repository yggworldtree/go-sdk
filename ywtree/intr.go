package ywtree

type IYWTListener interface {
	OnConnect(c *Engine)
	OnDisconnect(c *Engine)
	OnMessage(c *Engine, msg *MessageTopic)
	OnBroadcast(c *Engine, msg *MessageTopic)
}
