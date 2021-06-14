package ywtree

import (
	"github.com/yggworldtree/go-core/bean"
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
