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
	//hbtp.Debugf("Engine subs code:%d,err:%v,conts:%s", code, err, bts)
	return nil
}
func (c *Engine) UnSubTopic(pars []*bean.TopicPath) error {
	if pars == nil || len(pars) <= 0 {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("UnSubTopic")
	defer req.Close()
	err := req.Do(c.ctx, &bean.ClientUnSubTopic{
		Topics: pars,
	})
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return nil
}
func (c *Engine) PushTopic(pth *bean.TopicPath,
	data interface{}, hd ...interface{}) error {
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("PushTopic")
	defer req.Close()
	if pth != nil {
		req.SetArg("topicPath", pth.String())
	}
	err := req.Do(c.ctx, data, hd)
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return nil
}
func (c *Engine) GroupClients(org, name string) ([]*bean.GroupClients, error) {
	var ls []*bean.GroupClients
	pars := hbtp.Map{"fullPath": fmt.Sprintf("%s/%s", org, name)}
	err := c.doHbtpJson("GroupClients", pars, &ls)
	if err != nil {
		return nil, err
	}
	return ls, nil
}
func (c *Engine) HbtpGrpcRequest(pth *bean.CliGroupPath,
	control int32, cmd string, args ...url.Values) (*hbtp.Request, error) {
	req := c.newHbtpReq("GrpcClientReq")
	defer req.Close()
	req.SetArg("cliPath", pth.String())
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

func (c *Engine) CreateBucket(bucket string) error {
	if bucket == "" {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("CreateBucket")
	defer req.Close()
	req.SetArg("bucket", bucket)
	err := req.Do(c.ctx, nil)
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return nil
}
func (c *Engine) DeleteBucket(bucket string) error {
	if bucket == "" {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("DeleteBucket")
	defer req.Close()
	req.SetArg("bucket", bucket)
	err := req.Do(c.ctx, nil)
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return nil
}
func (c *Engine) SetBucketParam(bucket, key string, data []byte) error {
	if bucket == "" || key == "" {
		return errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("SetBucketParam")
	defer req.Close()
	req.SetArg("bucket", bucket)
	req.SetArg("key", key)
	err := req.Do(c.ctx, data)
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return nil
}
func (c *Engine) GetBucketParam(bucket, key string) ([]byte, error) {
	if bucket == "" || key == "" {
		return nil, errors.New("param err")
	}
	/*if len(data)>common.MaxTopicLen{
		return fmt.Errorf("topic data length out over:%d",common.MaxTopicLen)
	}*/
	req := c.newHbtpReq("GetBucketParam")
	defer req.Close()
	req.SetArg("bucket", bucket)
	req.SetArg("key", key)
	err := req.Do(c.ctx, nil)
	if err != nil {
		return nil, err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return nil, fmt.Errorf("server err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return req.ResBodyBytes(), nil
}
