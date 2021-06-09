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
	"github.com/yggworldtree/go-core/messages"
	"github.com/yggworldtree/go-core/utils"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

type Engine struct {
	info hbtpBean.ClientRegRes
	cfg  *Config
	ctx  context.Context
	cncl context.CancelFunc
	conn net.Conn
	htms time.Time
	htmr time.Time

	regd  bool
	sndch chan *clientBean.MessageBox
	rcvch chan *clientBean.MessageBox

	replylk sync.Mutex
	replymp map[string]messages.IReply
}

func NewEngine(ctx context.Context, cfg *Config) *Engine {
	c := &Engine{
		cfg:     cfg,
		sndch:   make(chan *clientBean.MessageBox, 100),
		rcvch:   make(chan *clientBean.MessageBox, 100),
		replymp: make(map[string]messages.IReply),
	}
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	return c
}
func (c *Engine) Ctx() context.Context {
	return c.ctx
}
func (c *Engine) Stop() {
	if c != nil && c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
	if c.sndch != nil {
		close(c.sndch)
		c.sndch = nil
	}
	if c.rcvch != nil {
		close(c.rcvch)
		c.rcvch = nil
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
			if err := c.runRead(); err != nil {
				logrus.Errorf("Client runRead err(end):%v", err)
				c.close()
			}
		}
	}()
	go func() {
		for !utils.EndContext(c.ctx) {
			c.runWrite()
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
	c.close()
	return nil
}
func (c *Engine) close() {
	conn := c.conn
	if conn != nil {
		conn.Close()
		c.conn = nil
		c.regd = false
	}
}
func (c *Engine) subs() {
	code, bts, err := c.DoHbtpString("SubTopic", &hbtpBean.ClientSubTopic{
		Topics: []string{"haha", "123123123"},
	})
	logrus.Debugf("Engine subs code:%d,err:%v,conts:%s", code, err, bts)
}
func (c *Engine) reg() error {
	c.regd = false
	req, err := c.NewHbtpReq("Reg", time.Second*5)
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
	info := &hbtpBean.ClientRegRes{}
	err = req.ResBodyJson(info)
	if err != nil {
		return err
	}
	if info.Id == "" || len(info.Token) < 32 {
		return fmt.Errorf("id code(%d) err:%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	c.info = *info
	c.conn = req.Conn(true)
	c.htms = time.Now()
	c.htmr = time.Now()
	c.regd = true
	c.subs()
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
			logrus.Infof("register runner suceess!id:%s", c.info.Id)
		}
	} else if time.Since(c.htms).Seconds() > 10 {
		c.htms = time.Now()
		messages.NewReplyCallback(c, clientBean.NewMessageBox(messages.MsgCmdHeart)).
			Ok(func(c messages.IEngine, m *clientBean.MessageBox) {
				logrus.Debugf("heart msg callback:%s!!!!!", m.Head.Id)
			}).
			Err(func(c messages.IEngine, errs error) {
				logrus.Debugf("heart msg callback errs:%v!!!!!", errs)
			}).Exec()
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
	conn := c.conn
	if !c.regd || conn == nil {
		time.Sleep(time.Millisecond)
		return nil
	}

	bts, err := utils.TcpRead(c.ctx, conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8d {
		logrus.Error("Client runRead 0x8d what????")
		return nil
	}
	bts, err = utils.TcpRead(c.ctx, conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8f {
		logrus.Error("Client runRead 0x8f what????")
		return nil
	}
	bts, err = utils.TcpRead(c.ctx, conn, 4)
	if err != nil {
		return err
	}
	hdln := uint(utils.BigByteToInt(bts))
	bts, err = utils.TcpRead(c.ctx, conn, 4)
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

	msg := &clientBean.MessageBox{}
	if hdln > 0 {
		bts, err = utils.TcpRead(c.ctx, conn, hdln)
		if err != nil {
			return err
		}
		err = json.Unmarshal(bts, &msg.Head)
		if err != nil {
			return nil
		}
	}
	if bodyln > 0 {
		bts, err = utils.TcpRead(c.ctx, conn, bodyln)
		if err != nil {
			return err
		}
		msg.Body = bts
	}
	bts, err = utils.TcpRead(c.ctx, conn, 2)
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
		time.Sleep(time.Millisecond)
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

	conn := c.conn
	if conn == nil {
		return
	}
	//logrus.Debugf("send msg:%s(%s)", msg.Head.Command, msg.Head.Id)
	hdln := utils.BigIntToByte(int64(len(hds)), 4)
	bodyln := utils.BigIntToByte(int64(len(msg.Body)), 4)
	conn.Write([]byte{0x8d, 0x8f})
	conn.Write(hdln)
	conn.Write(bodyln)
	if len(hds) > 0 {
		conn.Write(hds)
	}
	if len(msg.Body) > 0 {
		conn.Write(msg.Body)
	}
	conn.Write([]byte{0x8e, 0x8f})
}
func (c *Engine) runRecv() (rterr error) {
	defer func() {
		if err := recover(); err != nil {
			rterr = fmt.Errorf("recover:%v", rterr)
			logrus.Errorf("Engine run recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	msg, ok := <-c.rcvch
	if !ok {
		time.Sleep(time.Millisecond)
		return errors.New("rcvch is closed?")
	}
	if msg.Head == nil {
		return nil
	}
	if msg.Head.Command == messages.MsgCmdReply {
		mid := msg.Head.Args.GetString("mid")
		if mid == "" {
			return fmt.Errorf("recv msg-%s err: mid empty", msg.Head.Command)
		}
		c.replylk.Lock()
		e, ok := c.replymp[mid]
		delete(c.replymp, mid)
		c.replylk.Unlock()
		if !ok {
			return nil
		}
		fn := e.OkFun()
		if fn != nil {
			fn(c, msg)
		}
		return nil
	}

	fn, ok := mpCliFn[msg.Head.Command]
	if ok && fn != nil {
		fn(c, msg)
	} else {
		logrus.Debugf("Engine recv noExist msg-%s:%s", msg.Head.Command, string(msg.Body))
	}
	return nil
}
func (c *Engine) Send(cmd string, body []byte, args ...utils.Map) error {
	msg := clientBean.NewMessageBox(cmd)
	msg.Body = body
	if len(args) > 0 {
		msg.Head.Args = args[0]
	}
	return c.Sends(msg)
}
func (c *Engine) Sends(msg *clientBean.MessageBox) error {
	if msg == nil || msg.Head == nil || msg.Head.Command == "" {
		return errors.New("msg param err")
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
	return nil
}
func (c *Engine) SendReply(m *clientBean.MessageBox, stat string, body ...interface{}) error {
	if m == nil || m.Head == nil || m.Head.Id == "" {
		return errors.New("msg err")
	}
	msg := clientBean.NewMessageBox(messages.MsgCmdReply, utils.Map{
		"mid":    m.Head.Id,
		"status": stat,
	})
	if len(body) > 0 {
		msg.PutBody(body[0])
	}
	return c.Sends(msg)
}
func (c *Engine) SendForReply(e messages.IReply) error {
	msg := e.Message()
	if msg == nil || msg.Head == nil || msg.Head.Id != "" {
		return errors.New("msg id err")
	}
	err := c.Sends(msg)
	if err != nil {
		return err
	}
	c.replylk.Lock()
	defer c.replylk.Unlock()
	c.replymp[msg.Head.Id] = e
	return nil
}
func (c *Engine) RmReply(e messages.IReply) error {
	if e.Message().Head.Id != "" {
		return errors.New("msg id err")
	}
	c.replylk.Lock()
	defer c.replylk.Unlock()
	delete(c.replymp, e.Message().Head.Id)
	return nil
}
