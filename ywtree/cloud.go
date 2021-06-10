package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/yggworldtree/go-core/utils"
	"time"
)

const (
	RPCHostCode = 1
)

func (c *Engine) genTokens(req *hbtp.Request) string {
	hd := req.ReqHeader()
	return utils.Md5String(hd.Path + hd.RelIp + hd.Times + c.info.Token)
}
func (c *Engine) newHbtpReq(method string, tmots ...time.Duration) (*hbtp.Request, error) {
	return hbtp.NewDoRPCReq(c.cfg.Host, RPCHostCode, method, c.info.Id, c.genTokens, tmots...)
}
func (c *Engine) doHbtpJson(method string, in, out interface{}, hd ...map[string]interface{}) error {
	req, err := c.newHbtpReq(method)
	if err != nil {
		return err
	}
	return hbtp.DoJson(req, in, out, hd...)
}
func (c *Engine) doHbtpString(method string, in interface{}, hd ...hbtp.Mp) (int, []byte, error) {
	req, err := c.newHbtpReq(method)
	if err != nil {
		return 0, nil, err
	}
	return hbtp.DoString(req, in, hd...)
}
