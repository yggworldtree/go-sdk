package ywtree

import (
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/utils"
)

type MessageTopic struct {
	Id     string             `json:"id"`
	Path   *bean.TopicPath    `json:"path"`
	Sender *bean.CliGroupPath `json:"sender"`
	Head   []byte             `json:"head"`
	Body   []byte             `json:"body"`
	header *utils.Map
}

func (c *MessageTopic) Header() *utils.Map {
	if c.header == nil {
		c.header = utils.NewMaps(c.Head)
	}
	return c.header
}
