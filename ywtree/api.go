package ywtree

import (
	"errors"
	"fmt"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/yggworldtree/go-core/bean"
	"net/url"
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
func (c *Engine) PushTopic(pth *bean.TopicPath,
	data interface{}, pars ...*bean.TopicParam) error {
	if pth == nil {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	hd := hbtp.Map{"topicPath": pth.String()}
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
func (c *Engine) GroupClients(org, name string) ([]*bean.GroupClients, error) {
	pars := hbtp.Map{"fullPath": fmt.Sprintf("%s/%s", org, name)}
	var ls []*bean.GroupClients
	err := c.doHbtpJson("GroupClients", pars, &ls)
	if err != nil {
		return nil, err
	}
	return ls, nil
}
func (c *Engine) GrpcClientRequest(pth *bean.CliGroupPath,
	control int32, cmd string, args ...url.Values) (*hbtp.Request, error) {
	req := c.newHbtpReq("GrpcClient")
	defer req.Close()
	req.ReqHeader().Set("cliPath", pth.String())
	err := req.Do(c.ctx, nil)
	if err != nil {
		return nil, err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return nil, fmt.Errorf("server err(%d):%s",
			req.ResCode(), string(req.ResBodyBytes()))
	}
	rtq := hbtp.NewConnRequest(req.Conn(true), control).
		Command(cmd)
	if len(args) > 0 {
		rtq.Args(args[0])
	}
	return rtq, nil
}
