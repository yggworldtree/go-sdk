package ywtree

import (
	"errors"
	"fmt"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/yggworldtree/go-core/bean"
)

func (c *Engine) SubTopic(pars []*bean.TopicInfo) error {
	code, bts, err := c.doHbtpString("SubTopic", &bean.ClientSubTopic{
		Topics: pars,
	})
	if err != nil {
		return err
	}
	if code != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", code, string(bts))
	}
	//logrus.Debugf("Engine subs code:%d,err:%v,conts:%s", code, err, bts)
	return nil
}
func (c *Engine) PushTopic(pth *bean.TopicPath, data interface{}, pars ...*bean.TopicParam) error {
	if pth == nil {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	hd := hbtp.Mp{"topicPath": pth}
	if len(pars) > 0 {
		hd["type"] = pars[0].Type
	}
	code, bts, err := c.doHbtpString("PushTopic", data, hd)
	if err != nil {
		return err
	}
	if code != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", code, string(bts))
	}
	//logrus.Debugf("Engine subs code:%d,err:%v,conts:%s", code, err, bts)
	return nil
}
