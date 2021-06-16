package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/yggworldtree/go-core/utils"
	"net/url"
	"strconv"
	"time"
)

const (
	RPCHostCode = 10
)

func (c *Engine) newHbtpReq(method string, tmots ...time.Duration) *hbtp.Request {
	devid := c.info.Id
	times := strconv.FormatInt(time.Now().Unix(), 10)
	random := utils.RandomString(20)
	pars := url.Values{}
	pars.Set("devid", devid)
	pars.Set("times", times)
	pars.Set("random", random)
	sign := utils.Md5String(devid + random + times + c.info.Token)
	pars.Set("sign", sign)
	return hbtp.NewRequest(c.cfg.Host, RPCHostCode, tmots...).Command(method).Args(pars)
}
func (c *Engine) doHbtpJson(method string, in, out interface{}, hd ...hbtp.Map) error {
	req := c.newHbtpReq(method)
	return hbtp.DoJson(req, in, out, hd...)
}
func (c *Engine) doHbtpString(method string, in interface{}, hd ...hbtp.Map) (int32, []byte, error) {
	req := c.newHbtpReq(method)
	return hbtp.DoString(req, in, hd...)
}
