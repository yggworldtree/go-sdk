package ywtree

import (
	"context"
	"errors"
	"fmt"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"github.com/sirupsen/logrus"
	"github.com/yggworldtree/go-core/bean"
	"github.com/yggworldtree/go-core/messages"
	"github.com/yggworldtree/go-core/utils"
	"net"
	"net/url"
	"runtime/debug"
	"sync"
	"time"
)

type Engine struct {
	info bean.ClientRegRes
	cfg  *Config
	ctx  context.Context
	cncl context.CancelFunc
	conn net.Conn
	htms time.Time
	htmr time.Time

	regd  bool
	sndch chan *messages.MessageBox
	rcvch chan *messages.MessageBox

	lsr     IYWTListener
	replylk sync.Mutex
	replymp map[string]messages.IReply

	hbtpconf  hbtp.Config
	hbtpfnlk  sync.Mutex
	hbtpfns   map[int32]hbtp.ConnFun
	hbtpnotfn hbtp.ConnFun
}

func NewEngine(ctx context.Context, lsr IYWTListener, cfg *Config) *Engine {
	c := &Engine{
		cfg:      cfg,
		lsr:      lsr,
		sndch:    make(chan *messages.MessageBox, 100),
		rcvch:    make(chan *messages.MessageBox, 100),
		replymp:  make(map[string]messages.IReply),
		hbtpfns:  make(map[int32]hbtp.ConnFun),
		hbtpconf: hbtp.MakeConfig(),
	}
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	return c
}
func (c *Engine) Id() string {
	return c.info.Id
}
func (c *Engine) Alias() string {
	return c.info.Alias
}
func (c *Engine) SetListener(lsr IYWTListener) {
	c.lsr = lsr
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
func (c *Engine) Run() error {
	if c.cfg == nil {
		return errors.New("config is nil")
	}
	if c.cfg.Host == "" || c.cfg.Org == "" || c.cfg.Name == "" {
		return errors.New("config param is err")
	}
	if c.lsr == nil {
		return errors.New("listener is nil")
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
	if c.lsr != nil {
		c.lsr.OnDisconnect(c)
	}
}
func (c *Engine) reg() error {
	c.regd = false
	req := hbtp.NewRequest(c.cfg.Host, RPCHostCode, time.Second*10).Command("Reg")
	defer req.Close()
	if c.info.Alias == "" {
		c.info.Alias = c.cfg.Alias
	}
	err := req.Do(c.ctx, &bean.ClientRegInfo{
		Id:    c.info.Id,
		Org:   c.cfg.Org,
		Name:  c.cfg.Name,
		Alias: c.info.Alias,
	})
	if err != nil {
		return err
	}
	if req.ResCode() != hbtp.ResStatusOk {
		return fmt.Errorf("res code(%d) not ok:%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	info := &bean.ClientRegRes{}
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
			logrus.Infof("register runner suceess!id:%s,alias:%s", c.info.Id, c.info.Alias)
			if c.lsr != nil {
				c.lsr.OnConnect(c)
			}
		}
	} else if time.Since(c.htms).Seconds() > 10 {
		c.htms = time.Now()
		messages.NewReplyCallback(c, messages.NewMessageBox(messages.CmdHeart)).
			/*Ok(func(c messages.IEngine, m *messages.MessageBox) {
				logrus.Debugf("heart msg callback:%s!!!!!", m.Info.Id)
			}).*/
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

	bts, err := hbtp.TcpRead(c.ctx, conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8d {
		logrus.Error("Client runRead 0x8d what????")
		return nil
	}
	bts, err = hbtp.TcpRead(c.ctx, conn, 1)
	if err != nil {
		return err
	}
	if bts[0] != 0x8f {
		logrus.Error("Client runRead 0x8f what????")
		return nil
	}
	msg, err := messages.ReadMessageBox(c.ctx, conn, &c.hbtpconf)
	if err != nil {
		return err
	}
	bts, err = hbtp.TcpRead(c.ctx, conn, 2)
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
	c.conn.Write([]byte{0x8d, 0x8f})
	messages.WriteMessageBox(c.conn, msg)
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

	msg, ok := <-c.rcvch
	if !ok {
		time.Sleep(time.Millisecond)
		return errors.New("rcvch is closed?")
	}
	if msg.Info == nil {
		return nil
	}
	go c.onMsg(msg)
	return nil
}
func (c *Engine) onMsg(msg *messages.MessageBox) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Engine onMsg recover:%+v", err)
			logrus.Errorf("%s", string(debug.Stack()))
		}
	}()

	switch msg.Info.Command {
	case messages.CmdTopicGet:
		c.onTopicGet(msg)
	case messages.CmdNetConnect:
		c.onNetConnect(msg)
	case messages.CmdReply:
		mid := msg.Info.Args.Get("mid")
		if mid == "" {
			logrus.Debugf("recv msg-%s err: mid empty", msg.Info.Command)
			return
		}
		c.replylk.Lock()
		e, ok := c.replymp[mid]
		delete(c.replymp, mid)
		c.replylk.Unlock()
		if !ok {
			logrus.Debugf("recv msg-%s err: mid empty", msg.Info.Command)
			return
		}
		fn := e.OkFun()
		if fn != nil {
			fn(c, msg)
		}
	default:
		logrus.Debugf("Engine recv noExist msg-%s:%s", msg.Info.Command, string(msg.Body))
	}
}
func (c *Engine) Send(cmd string, body []byte, args ...url.Values) error {
	msg := messages.NewMessageBox(cmd, args...)
	msg.Body = body
	return c.Sends(msg)
}
func (c *Engine) Sends(msg *messages.MessageBox) error {
	if msg == nil || msg.Info == nil || msg.Info.Command == "" {
		return errors.New("msg param err")
	}
	if msg.Info.Id == "" {
		msg.Info.Id = utils.NewXid()
	}
	if msg.Info.Sender == "" {
		msg.Info.Sender = c.info.CliGroupPath().String()
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
func (c *Engine) SendReply(m *messages.MessageBox, stat string, body ...interface{}) error {
	if m == nil || m.Info == nil || m.Info.Id == "" {
		return errors.New("msg err")
	}
	pars := url.Values{}
	pars.Set("mid", m.Info.Id)
	pars.Set("status", stat)
	msg := messages.NewMessageBox(messages.CmdReply, pars)
	if len(body) > 0 {
		msg.PutBody(body[0])
	}
	return c.Sends(msg)
}
func (c *Engine) SendForReply(e messages.IReply) error {
	msg := e.Message()
	if msg == nil || msg.Info == nil || msg.Info.Id != "" {
		return errors.New("msg id err")
	}
	err := c.Sends(msg)
	if err != nil {
		return err
	}
	c.replylk.Lock()
	defer c.replylk.Unlock()
	c.replymp[msg.Info.Id] = e
	return nil
}
func (c *Engine) RmReply(e messages.IReply) error {
	if e.Message().Info.Id != "" {
		return errors.New("msg id err")
	}
	c.replylk.Lock()
	defer c.replylk.Unlock()
	delete(c.replymp, e.Message().Info.Id)
	return nil
}
