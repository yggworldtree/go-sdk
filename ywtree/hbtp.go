package ywtree

import (
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"net"
	"runtime/debug"
)

func (c *Engine) RegHbtpFun(control int32, fn hbtp.ConnFun) bool {
	c.hbtpfnlk.Lock()
	defer c.hbtpfnlk.Unlock()
	_, ok := c.hbtpfns[control]
	if ok || fn == nil {
		hbtp.Debugf("Engine RegFun err:control(%d) is exist", control)
		return false
	}
	c.hbtpfns[control] = fn
	return true
}
func (c *Engine) RegGrpcParamFun(control int32, fn interface{}) bool {
	return c.RegHbtpFun(control, hbtp.ParamFunHandle(fn))
}
func (c *Engine) RegGrpcGrpcFun(control int32, rpc hbtp.IRPCRoute) bool {
	return c.RegHbtpFun(control, hbtp.GrpcFunHandle(rpc))
}
func (c *Engine) netHandleConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			hbtp.Debugf("Engine netHandleConn recover:%+v", err)
			hbtp.Debugf("%s", string(debug.Stack()))
		}
	}()
	needclose := true
	defer func() {
		if needclose {
			conn.Close()
		}
	}()
	res, err := hbtp.ParseContext(c.ctx, conn, c.hbtpconf)
	if err != nil {
		hbtp.Debugf("Engine ParseContext recover:%+v", err)
		return
	}
	needclose = c.recoverCallMapfn(res)
}
func (c *Engine) HbtpNotFoundFun(fn hbtp.ConnFun) {
	c.hbtpnotfn = fn
}
func (c *Engine) recoverCallMapfn(res *hbtp.Context) (rt bool) {
	rt = false
	defer func() {
		if err := recover(); err != nil {
			rt = false
			hbtp.Debugf("Engine recoverCallMapfn recover:%+v", err)
			hbtp.Debugf("%s", string(debug.Stack()))
		}
	}()

	c.hbtpfnlk.Lock()
	fn, ok := c.hbtpfns[res.Control()]
	c.hbtpfnlk.Unlock()
	if ok && fn != nil {
		fn(res)
	} else if c.hbtpnotfn != nil {
		c.hbtpnotfn(res)
	} else {
		res.ResString(hbtp.ResStatusNotFound, "Not Found Control Function")
	}
	return res.IsOwn()
}
