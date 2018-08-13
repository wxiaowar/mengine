package mengine

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"mengine/json"
)

func (eg *Engine) encHandle(h HFunc, w http.ResponseWriter, r *http.Request) {
	bts, e := ioutil.ReadAll(r.Body)
	if e != nil {
		eg.encfail(w, fmt.Sprint("readbody error %v", e), r.URL.Path)
		return
	}

	if eg.Descrypt == nil {
		eg.encfail(w, "descrypt handle nil", r.URL.Path)
		return
	}

	ctx := &Context{
		Request: r,
		BodyRaw: bts,
	}

	decbts, e := eg.Descrypt(ctx)
	if e != nil {
		eg.encfail(w, "check descrypt error", r.URL.Path, e.Error())
		return
	}

	ctx.BodyRaw = decbts

	bmap := make(map[string]interface{}) // body按照json格式解析
	if len(ctx.BodyRaw) > 0 {
		if e := json.Unmarshal(ctx.BodyRaw, &bmap); e != nil {
			eg.encfail(w, fmt.Sprintf("read umarsh error %v", e), r.URL.Path)
			return
		}
	}

	eg.handleEnc(h, ctx, w)
}

func (eg *Engine) handleEnc(h HFunc, ctx *Context, w http.ResponseWriter) {
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

	rbts, e := json.Marshal(res)
	if e != nil {
		eg.encfail(w, fmt.Sprint("marshal write error : %v", e), ctx.Path())
		return
	}

	if eg.Encrypt == nil {
		eg.encfail(w, "encrypt handle nil", ctx.Path())
		return
	}

	rtbts := eg.Encrypt(rbts)

	_, e = w.Write(rtbts)
	if e != nil {
		eg.Notice("ok", -1, ctx.RemoteAddr(), ctx.RequestURI(), ctx.GetAuth(), string(ctx.BodyRaw), detail, string(rbts), fmt.Sprintf("write res error : %v", e))
		return
	}

	eg.Notice("ok", 0, ctx.RemoteAddr(), ctx.RequestURI(), ctx.GetAuth(), string(ctx.BodyRaw), detail, string(rbts))
}

func (eg *Engine) encfail(w http.ResponseWriter, detail, path string, args ...interface{}) {
	result := map[string]interface{}{
		"code": -1,
		"msg":  "internal",
	}

	if eg.IsDebug {
		result["detail"] = detail
	}

	eg.Notice("fail", -1, path, detail, args)
	res, _ := json.Marshal(result)
	if eg.Encrypt != nil {
		res = eg.Encrypt(res)
	}

	w.Write(res)
}
