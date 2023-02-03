package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

type Getter interface{ Get(key string) string }
type Setter interface{ Set(key, val string) }
type Deleter interface{ Del(key string) }

type GetSetter interface {
	Getter
	Setter
}

type Mapper interface {
	GetSetter
	Deleter
}

type Array []interface{}
type Json map[string]interface{}

func (js Json) IGet(key string) interface{} {
	if key == "" {
		return nil
	}
	return js[key]
}

func (js Json) ISet(key string, val interface{}) {
	js[key] = val
}

func (js Json) SetBool(key string, val bool) {
	js[key] = val
}

func (js Json) SetNumber(key string, val float64) {
	js[key] = val
}

func (js Json) SetArray(key string, val Array) {
	js[key] = val
}

func (js Json) SetObject(key string, val Json) {
	js[key] = val
}

func (js Json) SetCode(code int) {
	js.ISet("code", code)
}

func (js Json) SetMsg(msg string) {
	js.Set("msg", msg)
}

func (js Json) SetError(err error) {
	if err == nil {
		return
	}
	js.Set("error", err.Error())
}

func (js Json) Switch(key string) {
	if key == "" {
		return
	}

	if v, ok := js[key]; ok {
		switch vv := v.(type) {
		case bool:
			js[key] = !vv
		case string:
			b, err := strconv.ParseBool(vv)
			if err == nil {
				js[key] = !b
			} else {
				js[key] = vv == ""
			}
		case int, int8, int16, int32, int64:
			i := reflect.ValueOf(vv).Int()
			switch i {
			case 0:
				js[key] = true
			default:
				js[key] = false
			}
		case uint, uint8, uint16, uint32, uint64:
			i := reflect.ValueOf(vv).Uint()
			switch i {
			case 0:
				js[key] = true
			default:
				js[key] = false
			}
		case float32, float64:
			f := reflect.ValueOf(vv).Float()
			switch f {
			case 0:
				js[key] = true
			default:
				js[key] = false
			}
		default:
			js[key] = true
		}
		return
	}

	js[key] = true
}

func (js Json) Get(key string) string {
	if key == "" {
		return ""
	}

	if v, ok := js[key]; ok {
		if s, ok := v.(string); ok {
			return s
		} else {
			return fmt.Sprint(v)
		}
	}

	return ""
}

func (js Json) Set(key string, val string) {
	js[key] = val
}

func (js Json) Del(key string) {
	delete(js, key)
}

func (js Json) Write(p []byte) (int, error) {
	err := json.Unmarshal(p, &js)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (js Json) Message() string {
	return js.Get("msg")
}

func (js Json) Code() int {
	code, err := strconv.Atoi(js.Get("code"))
	if err != nil {
		return -1
	}
	return code
}

func (js Json) Err() error {
	e := js.Get("error")
	if e == "" {
		return nil
	}
	return errors.New(e)
}

func (js Json) Bytes() []byte {
	b, _ := json.Marshal(js)
	return b
}

func (js Json) String() string {
	b, _ := json.MarshalIndent(js, "", "\t")
	return string(b)
}

func (js Json) ResponseTo(ctx context.Context, w http.ResponseWriter) {
	select {
	case <-ctx.Done():
	default:
		// hack for using defer
		if err := recover(); err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js.Bytes())
	}
}
