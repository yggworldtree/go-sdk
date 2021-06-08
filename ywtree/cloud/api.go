package cloud

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"time"
)

const (
	RPCHostCode = 1
)

func NewReq(host, method string, tmots ...time.Duration) (*hbtp.Request, error) {
	return hbtp.NewDoRPCReq(host, RPCHostCode, method, "", nil, tmots...)
}
func DoJson(host, method string, in, out interface{}, hd ...map[string]interface{}) error {
	req, err := NewReq(host, method)
	if err != nil {
		return err
	}
	return hbtp.DoJson(req, in, out, hd...)
}
func DoString(host, method string, in interface{}, hd ...hbtp.Mp) (int, []byte, error) {
	req, err := NewReq(host, method)
	if err != nil {
		return 0, nil, err
	}
	return hbtp.DoString(req, in, hd...)
}
