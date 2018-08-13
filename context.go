package mengine

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mengine/json"
	"mpkg/convert"
	"net/http"
	"os"
	"reflect"
	"strings"
)

type Context struct {
	Body    map[string]interface{}
	BodyRaw []byte

	Request *http.Request
	Uid     int64
}

func (ctx *Context) GetParam(name string) string {
	return ctx.Request.URL.Query().Get(name)
}

func (ctx *Context) IP() string {
	ips := strings.Split(ctx.Request.Header.Get("X-Forwarded-For"), ",")
	if len(ips[0]) > 3 {
		return ips[0]
	} else {
		addr := strings.Split(ctx.Request.RemoteAddr, ":")
		return addr[0]
	}
}

func (ctx *Context) RemoteAddr() string {
	return ctx.Request.RemoteAddr
}

func (ctx *Context) Cookie(name string) (*http.Cookie, error) {
	return ctx.Request.Cookie(name)
}

func (ctx *Context) GetAuth() string {
	c, e := ctx.Request.Cookie("auth")
	if e != nil {
		return "auth=''"
	}

	return c.Value
}

func (ctx *Context) Path() string {
	return ctx.Request.URL.Path
}

func (ctx *Context) RequestURI() string {
	return ctx.Request.RequestURI
}

func (ctx *Context) Host() string {
	return ctx.Request.Host
}

func (ctx *Context) Header() http.Header {
	return ctx.Request.Header
}

//检查Body中的字段是否齐全
func (ctx *Context) EnsureBody(keys ...string) (string, bool) {
	for _, key := range keys {
		if _, ok := ctx.Body[key]; !ok {
			return key, false
		}
	}
	return "", true
}

// 兼容 科学计数法float64类型
func (ctx *Context) DecodeJson(v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(ctx.BodyRaw))
	decoder.UseNumber()
	return decoder.Decode(v)
}

//带默认值的解析
func (ctx *Context) ParseOpt(params ...interface{}) error {
	if len(params)%3 != 0 {
		return errors.New("params count invalid")
	}
	for i := 0; i < len(params); i += 3 {
		key := convert.ToString(params[i])
		v, ok := ctx.Body[key]
		var e error
		switch ref := params[i+1].(type) {
		case *string:
			if ok {
				*ref = convert.ToString(v)
			} else {
				*ref = convert.ToString(params[i+2])
			}
		case *float64:
			if ok {
				*ref, e = convert.ToFloat64(v)
			} else {
				*ref, e = convert.ToFloat64(params[i+2])
			}
		case *int:
			if ok {
				*ref, e = convert.ToInt(v)
			} else {
				*ref, e = convert.ToInt(params[i+2])
			}
		case *int8:
			if ok {
				*ref, e = convert.ToInt8(v)
			} else {
				*ref, e = convert.ToInt8(params[i+2])
			}
		case *int16:
			if ok {
				*ref, e = convert.ToInt16(v)
			} else {
				*ref, e = convert.ToInt16(params[i+2])
			}
		case *int32:
			if ok {
				*ref, e = convert.ToInt32(v)
			} else {
				*ref, e = convert.ToInt32(params[i+2])
			}
		case *int64:
			if ok {
				*ref, e = convert.ToInt64(v)
			} else {
				*ref, e = convert.ToInt64(params[i+2])
			}
		case *uint:
			if ok {
				*ref, e = convert.ToUint(v)
			} else {
				*ref, e = convert.ToUint(params[i+2])
			}
		case *uint8:
			if ok {
				*ref, e = convert.ToUint8(v)
			} else {
				*ref, e = convert.ToUint8(params[i+2])
			}
		case *uint16:
			if ok {
				*ref, e = convert.ToUint16(v)
			} else {
				*ref, e = convert.ToUint16(params[i+2])
			}
		case *uint32:
			if ok {
				*ref, e = convert.ToUint32(v)
			} else {
				*ref, e = convert.ToUint32(params[i+2])
			}
		case *uint64:
			if ok {
				*ref, e = convert.ToUint64(v)
			} else {
				*ref, e = convert.ToUint64(params[i+2])
			}
		case *bool:
			if ok {
				*ref, e = convert.ToBool(v)
			} else {
				*ref, e = convert.ToBool(params[i+2])
			}
		case *[]string:
			if ok {
				*ref, e = convert.ToStringSlice(v)
			} else {
				*ref = params[i+2].([]string)
			}
		case *[]int64:
			if ok {
				*ref, e = convert.ToInt64Slice(v)
			} else {
				*ref = params[i+2].([]int64)
			}
		case *[]uint32:
			if ok {
				*ref, e = convert.ToUint32Slice(v)
			} else {
				*ref = params[i+2].([]uint32)
			}
		case *map[string]interface{}:
			if ok {
				switch m := v.(type) {
				case map[string]interface{}:
					*ref = m
				default:
					e = errors.New(fmt.Sprintf("%v is not map[string]iterface{}", key, reflect.TypeOf(v)))
				}
			} else {
				*ref = params[i+2].(map[string]interface{})
			}
		case *[]interface{}:
			if ok {
				switch m := v.(type) {
				case []interface{}:
					*ref = m
				default:
					e = errors.New("value is not []iterface{}")
				}
			} else {
				*ref = params[i+2].([]interface{})
			}
		case *interface{}:
			if ok {
				*ref = v
			} else {
				*ref = params[i+2]
			}
		default:
			return errors.New(fmt.Sprintf("unknown type %v ", key, reflect.TypeOf(ref)))
		}
		if e != nil {
			return errors.New(fmt.Sprintf("parse [%v] error:%v", key, e.Error()))
		}
	}
	return nil
}

//不带默认值的解析
func (ctx *Context) Parse(params ...interface{}) error {
	if len(params)%2 != 0 {
		return errors.New("params count must be even")
	}
	for i := 0; i < len(params); i += 2 {
		key := convert.ToString(params[i])
		if v, ok := ctx.Body[key]; ok {
			var e error
			switch ref := params[i+1].(type) {
			case *string:
				*ref = convert.ToString(v)
			case *float64:
				*ref, e = convert.ToFloat64(v)
			case *int:
				*ref, e = convert.ToInt(v)
			case *int8:
				*ref, e = convert.ToInt8(v)
			case *int16:
				*ref, e = convert.ToInt16(v)
			case *int32:
				*ref, e = convert.ToInt32(v)
			case *int64:
				*ref, e = convert.ToInt64(v)
			case *uint:
				*ref, e = convert.ToUint(v)
			case *uint8:
				*ref, e = convert.ToUint8(v)
			case *uint16:
				*ref, e = convert.ToUint16(v)
			case *uint32:
				*ref, e = convert.ToUint32(v)
			case *uint64:
				*ref, e = convert.ToUint64(v)
			case *bool:
				*ref, e = convert.ToBool(v)
			case *[]string:
				*ref, e = convert.ToStringSlice(v)
			case *[]int64:
				*ref, e = convert.ToInt64Slice(v)
			case *[]uint32:
				*ref, e = convert.ToUint32Slice(v)
			case *map[string]interface{}:
				switch m := v.(type) {
				case map[string]interface{}:
					*ref = m
				default:
					e = errors.New("value is not map[string]iterface{}")
				}
			case *[]interface{}:
				switch m := v.(type) {
				case []interface{}:
					*ref = m
				default:
					e = errors.New("value is not []iterface{}")
				}
			case *interface{}:
				*ref = v
			default:
				return errors.New(fmt.Sprintf("unknown type %v ", reflect.TypeOf(ref)))
			}
			if e != nil {
				return errors.New(fmt.Sprintf("parse [%v] error:%v", key, e.Error()))
			}
		} else {
			return errors.New(fmt.Sprintf("%v not provided", key))
		}
	}
	return nil
}

func (ctx *Context) PostFile(param string, filename string) (e error) {
	// hr.request.ParseMultipartForm(1024 * 1024 * 10)
	file, _, e := ctx.Request.FormFile(param)
	if e != nil {
		return errors.New(fmt.Sprintf("FormFile: %v", e.Error()))
	}

	defer file.Close()

	bytes, e := ioutil.ReadAll(file)
	if e != nil {
		return errors.New(fmt.Sprintf("ReadAll: %v", e.Error()))
	}
	e = ioutil.WriteFile(filename, bytes, os.ModePerm)
	return
}

//获取HTTP post的详细信息
func (ctx *Context) PostFileInfo(param string) (bytes []byte, filename string, e error) {
	// hr.request.ParseMultipartForm(1024 * 1024 * 10)
	file, handler, e := ctx.Request.FormFile(param)
	if e != nil {
		e = errors.New(fmt.Sprintf("FormFile: %v", e.Error()))
		return
	}
	defer file.Close()
	filename = handler.Filename
	bytes, e = ioutil.ReadAll(file)
	if e != nil {
		e = errors.New(fmt.Sprintf("ReadAll: %v", e.Error()))
		return
	}
	// e = ioutil.WriteFile(filename, bytes, os.ModePerm)
	return
}

//参照prase 写的getparams。
func (ctx *Context) GetParams(params ...interface{}) error {
	if len(params)%2 != 0 {
		return errors.New("params count must be odd")
	}
	for i := 0; i < len(params); i += 2 {
		key := convert.ToString(params[i])
		if v := ctx.Request.URL.Query().Get(key); v != "" {
			var e error
			switch ref := params[i+1].(type) {
			case *string:
				*ref = convert.ToString(v)
			case *float64:
				*ref, e = convert.ToFloat64(v)
			case *int:
				*ref, e = convert.ToInt(v)
			case *int8:
				*ref, e = convert.ToInt8(v)
			case *int16:
				*ref, e = convert.ToInt16(v)
			case *int32:
				*ref, e = convert.ToInt32(v)
			case *int64:
				*ref, e = convert.ToInt64(v)
			case *uint:
				*ref, e = convert.ToUint(v)
			case *uint8:
				*ref, e = convert.ToUint8(v)
			case *uint16:
				*ref, e = convert.ToUint16(v)
			case *uint32:
				*ref, e = convert.ToUint32(v)
			case *uint64:
				*ref, e = convert.ToUint64(v)
			case *bool:
				*ref, e = convert.ToBool(v)
			case *map[string]interface{}:
				e = errors.New("do not support map[string]iterface{}")
			case *[]interface{}:
				e = errors.New("do not support []iterface{}")
			case *[]string:
				*ref, e = convert.ToStringSlice(v)
			case *[]uint32:
				*ref, e = convert.ToUint32Slice(v)
			case *interface{}:
				*ref = v
			default:
				return errors.New(fmt.Sprintf("unknown type %v ", reflect.TypeOf(ref)))
			}
			if e != nil {
				return errors.New(fmt.Sprintf("parse1111 [%v] error:%v", key, e.Error()))
			}
		} else {
			return errors.New(fmt.Sprintf("%v not provided value", key))
		}
	}
	return nil
}

func (ctx *Context) GetParamOpt(params ...interface{}) error {
	if len(params)%3 != 0 {
		return errors.New("params count invalid")
	}
	for i := 0; i < len(params); i += 3 {
		key := convert.ToString(params[i])
		v := ctx.Request.URL.Query().Get(key)
		var e error
		switch ref := params[i+1].(type) {
		case *string:
			if v != "" {
				*ref = convert.ToString(v)
			} else {
				*ref = convert.ToString(params[i+2])
			}
		case *float64:
			if v != "" {
				*ref, e = convert.ToFloat64(v)
			} else {
				*ref, e = convert.ToFloat64(params[i+2])
			}
		case *int:
			if v != "" {
				*ref, e = convert.ToInt(v)
			} else {
				*ref, e = convert.ToInt(params[i+2])
			}
		case *int8:
			if v != "" {
				*ref, e = convert.ToInt8(v)
			} else {
				*ref, e = convert.ToInt8(params[i+2])
			}
		case *int16:
			if v != "" {
				*ref, e = convert.ToInt16(v)
			} else {
				*ref, e = convert.ToInt16(params[i+2])
			}
		case *int32:
			if v != "" {
				*ref, e = convert.ToInt32(v)
			} else {
				*ref, e = convert.ToInt32(params[i+2])
			}
		case *int64:
			if v != "" {
				*ref, e = convert.ToInt64(v)
			} else {
				*ref, e = convert.ToInt64(params[i+2])
			}
		case *uint:
			if v != "" {
				*ref, e = convert.ToUint(v)
			} else {
				*ref, e = convert.ToUint(params[i+2])
			}
		case *uint8:
			if v != "" {
				*ref, e = convert.ToUint8(v)
			} else {
				*ref, e = convert.ToUint8(params[i+2])
			}
		case *uint16:
			if v != "" {
				*ref, e = convert.ToUint16(v)
			} else {
				*ref, e = convert.ToUint16(params[i+2])
			}
		case *uint32:
			if v != "" {
				*ref, e = convert.ToUint32(v)
			} else {
				*ref, e = convert.ToUint32(params[i+2])
			}
		case *uint64:
			if v != "" {
				*ref, e = convert.ToUint64(v)
			} else {
				*ref, e = convert.ToUint64(params[i+2])
			}
		case *bool:
			if v != "" {
				*ref, e = convert.ToBool(v)
			} else {
				*ref, e = convert.ToBool(params[i+2])
			}
		case *[]string:
			if v != "" {
				*ref, e = convert.ToStringSlice(v)
			} else {
				*ref = params[i+2].([]string)
			}
		case *[]uint32:
			if v != "" {
				*ref, e = convert.ToUint32Slice(v)
			} else {
				*ref = params[i+2].([]uint32)
			}
		case *map[string]interface{}:
			e = errors.New("do not support map[string]iterface{}")
		case *[]interface{}:
			e = errors.New("do not support []iterface{}")
		default:
			return errors.New(fmt.Sprintf("unknown type %v ", key, reflect.TypeOf(ref)))
		}
		if e != nil {
			return errors.New(fmt.Sprintf("parse [%v] error:%v", key, e.Error()))
		}
	}
	return nil
}
