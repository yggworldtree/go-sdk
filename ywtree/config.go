package ywtree

import "github.com/yggworldtree/go-core/bean"

type Config struct {
	Host      string `json:"host"`
	Secret    string `json:"secret"`
	Org       string `json:"org"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Frequency string `json:"frequency"`
	Subs      []bean.TopicSubInfo
	Pushs     []bean.TopicPushInfo
	Sign      string `json:"sign"`
}
