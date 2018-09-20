/*
简易的Http框架，以json为传输格式。

错误返回值：
	{
		"code":2001,	// 错误码
		"msg":"参数错误",	// 客户端显示的错误原因
		"res": {        // 处理结果
		}
	}
*/
package mengine

import (
	oscontext "context"
	"fmt"
	"github.com/wxiaowar/mengine/json"
	"github.com/wxiaowar/mlog"
	"io/ioutil"
	"net/http"
	"time"
)

type HFunc func(c *Context, res map[string]interface{}) Error

type EngionOption struct {
	Addr    string
	IsDebug bool

	CheckWhiteList func(ctx *Context) (e error)             // 检测是IP白名单
	CheckIntegrity func(ctx *Context) (e error)             // 验证完整性
	Descrypt       func(ctx *Context) (bts []byte, e error) // 解密验证
	Encrypt        func(bts []byte) (edbts []byte)          // 加密
}

type Engine struct {
	*EngionOption
	Router
	*mlog.MLog
	svr *http.Server
}

//
func NewEngine(opt *EngionOption, log *mlog.MLog, r Router) *Engine {
	eg := &Engine{
		EngionOption: opt,
		MLog:         log,
		Router:       r,
	}

	// TODO pool reduice gc count
	//eg.pool.New = func() interface{} {
	//	return &Context{}
	//}
	return eg
}

// ServeHTTP conforms to the http.Handler interface.
func (eg *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 跨域问题
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "x-requested-with")
	w.Header().Add("Access-Control-Allow-Headers", "Cookie")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization")
	w.Header().Add("Access-Control-Allow-Headers", "auth")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	h, b := eg.Rout(req.URL.Path)
	if !b {
		eg.comfail(w, "invalid path", req.URL.Path)
		return
	}

	itype := req.URL.Path[1]
	switch itype {
	case 't':
		eg.trustHandle(h, w, req)
		break

	case 'i':
		eg.itgHandle(h, w, req)
		break

	case 'x':
		eg.encHandle(h, w, req)
		break

	default:
		eg.comfail(w, "invalid itype", req.URL.Path)
	}
}

//
func (eg *Engine) Run() error {
	s := &http.Server{
		Addr:           eg.Addr,
		Handler:        eg,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 0,
	}

	eg.svr = s

	return s.ListenAndServe()
}

// ShutDown graceful shutdown
func (eg *Engine) ShutDown(duration time.Duration) error {
	if eg.svr != nil {
		ctx, _ := oscontext.WithTimeout(oscontext.Background(), duration*time.Second)
		return eg.svr.Shutdown(ctx)
	}

	return nil
}

//
func (eg *Engine) RunTLS(addr, certFile, keyFile string) error {
	s := &http.Server{
		Addr:           eg.Addr,
		Handler:        eg,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 0,
	}

	eg.svr = s
	return s.ListenAndServeTLS(certFile, keyFile)
}

//
func readbody(r *http.Request) (bts []byte, bmap map[string]interface{}, e error) {
	bts, e = ioutil.ReadAll(r.Body)
	if e != nil {
		return
	}

	bmap = make(map[string]interface{})
	if len(bts) <= 0 { // len <= 0, no need unmarshal
		return
	}

	// body按照json格式解析
	if e = json.Unmarshal(bts, &bmap); e != nil {
		return
	}

	return bts, bmap, nil
}

// local server
func (eg *Engine) trustHandle(h HFunc, w http.ResponseWriter, r *http.Request) {
	bts, bmap, e := readbody(r)
	if e != nil {
		eg.comfail(w, fmt.Sprint("readbody error %v", e), r.URL.Path)
		return
	}

	if eg.CheckWhiteList == nil { // 本地白名单检测
		eg.comfail(w, "check white handle nil", r.URL.Path)
		return
	}

	if len(r.URL.Path) < 3 {
		eg.comfail(w, "invalid url formate", r.URL.Path)
		return
	}

	// TODO use sync.pool reduice gc
	ctx := &Context{
		Request: r,
		Body:    bmap,
		BodyRaw: bts,
	}

	if e = eg.CheckWhiteList(ctx); e != nil {
		eg.authfail(w, fmt.Sprintf("check white error %v", e), r.URL.Path)
		return
	}

	eg.handle(h, ctx, w)
}

//
func (eg *Engine) itgHandle(h HFunc, w http.ResponseWriter, r *http.Request) {
	bts, bmap, e := readbody(r)
	if e != nil {
		eg.comfail(w, fmt.Sprintf("readbody error %v", e), r.URL.Path)
		return
	}

	if eg.CheckIntegrity == nil {
		eg.comfail(w, "check integrity handle nil", r.URL.Path)
		return
	}

	// TODO use sync.pool reduice gc
	ctx := &Context{
		Request: r,
		Body:    bmap,
		BodyRaw: bts,
	}

	if e := eg.CheckIntegrity(ctx); e != nil {
		eg.authfail(w, fmt.Sprintf("invalid integrity msg %v", e), r.URL.Path)
		return
	}

	eg.handle(h, ctx, w)
}

//
func (eg *Engine) handle(h HFunc, ctx *Context, w http.ResponseWriter) {
	res := make(map[string]interface{})
	se := h(ctx, res)

	var (
		code   int32  = 0
		msg    string = "success"
		detail string = ""
	)

	if se != nil {
		code = se.Code()
		msg = se.Msg()
		detail = se.Detail()
	}

	res["code"] = code
	res["msg"] = msg
	res["tm"] = time.Now().Unix()

	if eg.IsDebug {
		res["detail"] = detail
	}

	rbts, err := json.Marshal(res)
	if err != nil {
		eg.comfail(w, fmt.Sprint("marshal write error : %v", err), ctx.Path())
		return
	}

	_, err = w.Write(rbts)
	if err != nil {
		eg.Error().Str("status", "ok").
			Int("code", -1).
			Str("remote", ctx.RemoteAddr()).
			Str("uri", ctx.RequestURI()).
			Str("auth", ctx.GetAuth()).
			Str("body", string(ctx.BodyRaw)).
			Str("detail", detail).Msg(err.Error())
		return
	}

	eg.Info().Str("status", "ok").
		Int("code", 0).
		Str("remote", ctx.RemoteAddr()).
		Str("uri", ctx.RequestURI()).
		Str("auth", ctx.GetAuth()).
		Str("body", string(ctx.BodyRaw)).
		Str("return", string(rbts))
}

func (eg *Engine) authfail(w http.ResponseWriter, detail, path string) {
	eg.fail(w, -2, detail, path)
}

func (eg *Engine) comfail(w http.ResponseWriter, detail, path string) {
	eg.fail(w, -1, detail, path)
}

func (eg *Engine) fail(w http.ResponseWriter, code int, detail, path string) {
	result := map[string]interface{}{
		"code": code,
		"msg":  "internal",
	}

	if eg.IsDebug {
		result["detail"] = detail
	}

	eg.Error().Str("status", "fail").
		Int("code", code).
		Str("path", path).Str("detail", detail)
	res, _ := json.Marshal(result)
	w.Write(res)
}
