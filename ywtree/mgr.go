package ywtree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean/clientBean"
	"github.com/yggworldtree/go-core/bean/hbtpBean"
	"github.com/yggworldtree/go-core/common"
	"github.com/yggworldtree/go-core/utils"
	"github.com/yggworldtree/go-sdk/ywtree/cloud"
	"net"
	"runtime/debug"
	"time"
)

type Engine struct {
	id   string
	cfg  *Config
	ctx  context.Context
	cncl context.CancelFunc
	conn net.Conn
	htms time.Time
	htmr time.Time

	regd  bool
	sndch chan *clientBean.MessageBox
	rcvch chan *clientBean.MessageBox
}

func NewEngine(cfg *Config) *Engine {
	c := &Engine{
		cfg:   cfg,
		sndch: make(chan *clientBean.MessageBox, 100),
		rcvch: make(chan *clientBean.MessageBox, 100),
	}
	c.ctx, c.cncl = context.WithCancel(context.Background())
	return c
}
func (c *Engine) Stop() {
	if c != nil && c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
}
func (c *Engine) Run() (rterr error) {
	if c.cfg == nil {
		return errors.New("config is nil")
	}
	if c.cfg.Host == "" || c.cfg.Org == "" || c.cfg.Name == "" {
		return errors.New("config param is err")
	}

	go func() {
		for !utils.EndContext(c.ctx) {
			if c.regd && c.conn != nil {
				if err := c.runRead(); err != nil {
					logrus.Errorf("Client runRead err(end):%v", err)
					c.close()
				}
			} else {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
	go func() {
		for !utils.EndContext(c.ctx) {
			if c.regd && c.conn != nil {
				c.runWrite()
			} else {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
	go func() {
		for !utils.EndContext(c.ctx) {
			err := c.runRecv()
			if err != nil {
				logrus.Errorf("Engine runRecv err(end):%v", err)
			}
		}
	}()
	for !utils.EndContext(c.ctx) {
		err := c.run()
		if err != nil {
			logrus.Errorf("Engine run err(end):%v", err)
			//return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}
func (c *Engine) close() {
	conn := c.conn
	if conn != nil {
		conn.Close()
		c.conn = nil
		c.regd = false
	}
	/*if c.sndch!=nil{
		close(c.sndch)
		c.sndch=nil
	}
	if c.rcvch!=nil{
		close(c.rcvch)
		c.rcvch=nil
	}*/
}
func (c *Engine) reg() error {
	c.regd = false
	req, err := cloud.NewReq(c.cfg.Host, "Reg", time.Second*5)
	if err != nil {
		return err
	}
	defer req.Close()
	err = req.Do(c.ctx, &hbtpBean.ClientRegInfo{
		Org:  c.cfg.Org,
		Name: c.cfg.Name,
	})
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("res code(%d) not ok:%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	ids := string(req.ResBodyBytes())
	if ids == "" {
		return fmt.Errorf("id code(%d) err:%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	c.id = ids
	//c.sndch = make(chan *clientBean.MessageBox,100)
	//c.rcvch = make(chan *clientBean.MessageBox,100)
	c.conn = req.Conn(true)
	c.htms = time.Now()
	c.htmr = time.Now()
	c.regd = true
	return nil
}
func (c *Engine) run() (rterr error) {
	defer func() {
		if err := recover(); err != nil {
			rterr = fmt.Errorf("recover:%v", rterr)
			logrus.Errorf("Engine run recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	if c.conn == nil {
		if err := c.reg(); err != nil {
			/*switch err.(type) {
			case *net.OpError:
				operr := err.(*net.OpError)
				if !strings.Contains(comm.MainCfg.ServAddr, "server:") && !operr.Timeout() {
					logrus.Errorf("register server(%s) failed:%v", comm.MainCfg.ServAddr, err)
				}
			default:
				logrus.Errorf("register servers(%s) failed:%v", comm.MainCfg.ServAddr, err)
			}*/
			logrus.Errorf("register servers(%s) failed:%v", c.cfg.Host, err)
			time.Sleep(time.Second * 3)
		} else {
			logrus.Infof("register runner suceess!id:%s", c.id)
		}
	} else if time.Since(c.htms).Seconds() > 10 {
		c.Send("msg/heart", nil)
		c.htms = time.Now()
	} else if time.Since(c.htmr).Seconds() > 32 {
		c.close()
		time.Sleep(time.Second * 1)
	}

	return nil
}

func (c *Engine) runRead() error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Engine runRead recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	bts, err := utils.TcpRead(c.ctx, c.conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8d {
		logrus.Error("Client runRead 0x8d what????")
		return nil
	}
	bts, err = utils.TcpRead(c.ctx, c.conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8f {
		logrus.Error("Client runRead 0x8f what????")
		return nil
	}
	bts, err = utils.TcpRead(c.ctx, c.conn, 4)
	if err != nil {
		return err
	}
	hdln := uint(utils.BigByteToInt(bts))
	bts, err = utils.TcpRead(c.ctx, c.conn, 4)
	if err != nil {
		return err
	}
	bodyln := uint(utils.BigByteToInt(bts))

	if hdln > common.MaxCliHeadLen {
		logrus.Errorf("Client runRead hdln out:%d/%d", hdln, common.MaxCliHeadLen)
		return errors.New("hdln out max")
	}
	if bodyln > common.MaxCliBodyLen {
		logrus.Errorf("Client runRead bodyln out:%d/%d", bodyln, common.MaxCliBodyLen)
		return errors.New("bodyln out max")
	}

	msg := clientBean.NewMessageBox()
	if hdln > 0 {
		bts, err = utils.TcpRead(c.ctx, c.conn, hdln)
		if err != nil {
			return err
		}
		err = json.Unmarshal(bts, &msg.Head)
		if err != nil {
			return nil
		}
	}
	if bodyln > 0 {
		bts, err = utils.TcpRead(c.ctx, c.conn, bodyln)
		if err != nil {
			return err
		}
		msg.Body = bts
	}
	bts, err = utils.TcpRead(c.ctx, c.conn, 2)
	if err != nil {
		return err
	}
	if c.rcvch != nil && bts[0] == 0x8e && bts[1] == 0x8f {
		c.rcvch <- msg
		c.htmr = time.Now()
	}
	return nil
}
func (c *Engine) runWrite() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Engine runWrite recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	if c.sndch == nil {
		return
	}
	msg, ok := <-c.sndch
	if !ok {
		return
	}
	var hds []byte
	if msg.Head != nil {
		bts, err := json.Marshal(msg.Head)
		if err != nil {
			logrus.Errorf("Client runWrite json err:%+v", err)
			return
		}
		hds = bts
	}

	if c.conn == nil {
		return
	}
	logrus.Debugf("send msg:%s(%s)", msg.Head.Command, msg.Head.Id)
	hdln := utils.BigIntToByte(int64(len(hds)), 4)
	bodyln := utils.BigIntToByte(int64(len(msg.Body)), 4)
	c.conn.Write([]byte{0x8d, 0x8f})
	c.conn.Write(hdln)
	c.conn.Write(bodyln)
	if len(hds) > 0 {
		c.conn.Write(hds)
	}
	if len(msg.Body) > 0 {
		c.conn.Write(msg.Body)
	}
	c.conn.Write([]byte{0x8e, 0x8f})
}
func (c *Engine) runRecv() (rterr error) {
	defer func() {
		if err := recover(); err != nil {
			rterr = fmt.Errorf("recover:%v", rterr)
			logrus.Errorf("Engine run recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	select {
	case msg := <-c.rcvch:
		if msg.Head != nil {
			logrus.Debugf("Engine recv msg-%s:%s", msg.Head.Command, string(msg.Body))
			/*if fn, ok := mpCliFn[msg.Head.Command]; ok && fn != nil {
				fn(c, msg)
			}*/
		}
	default:
		time.Sleep(time.Millisecond)
	}
	return nil
}
func (c *Engine) Send(cmd string, body []byte, args ...utils.Map) {
	msg := clientBean.NewMessageBox()
	msg.Head.Command = cmd
	msg.Body = body
	if len(args) > 0 {
		msg.Head.Args = args[0]
	}
	c.Sends(msg)
}
func (c *Engine) Sends(msg *clientBean.MessageBox) {
	if msg == nil || msg.Head == nil || msg.Head.Command == "" {
		return
	}
	if msg.Head.Id == "" {
		msg.Head.Id = utils.NewXid()
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Sends recover:%+v", err)
			}
		}()
		c.sndch <- msg
	}()
}
func (c *Engine) SendReply(m *clientBean.MessageBox, stat string) {
	if m == nil || m.Head == nil || m.Head.Id == "" {
		return
	}
	msg := clientBean.NewMessageBox()
	msg.Head.Command = "msg/reply"
	msg.PutBody(&clientBean.MessageReply{
		Id:     m.Head.Id,
		Status: stat,
	})
	c.Sends(msg)
}
