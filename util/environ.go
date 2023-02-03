package util

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func InspectEnv(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return errors.New("expect struct pointer type")
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return errors.New("expect struct pointer type")
	}

	var envmap = make(map[string]string)
	for _, env := range os.Environ() {
		i := strings.Index(env, "=")
		if i < 0 {
			continue
		}
		key, value := env[:i], env[i+1:]
		envmap[key] = value
	}

	for i, vType := 0, v.Type(); i < vType.NumField(); i++ {
		vf := vType.Field(i)
		if !strings.ContainsAny(vf.Name[:1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			continue
		}

		vtag := vf.Tag.Get("env")
		if vtag == "-" {
			continue
		}

		tag := parseTag(vtag)
		name := tag["name"]
		if name == "" {
			name = strings.ToUpper(vf.Name)
		}

		value, ok := envmap[name]
		if !ok {
			dv, ok := tag["default"]
			if !ok {
				return errors.New("no value found for field " + vf.Name)
			}
			value = dv
		}

		switch vf.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iv, _ := strconv.ParseInt(value, 10, 64)
			v.Field(i).SetInt(iv)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			iv, _ := strconv.ParseUint(value, 10, 64)
			v.Field(i).SetUint(iv)
		case reflect.Float32, reflect.Float64:
			fv, _ := strconv.ParseFloat(value, 64)
			v.Field(i).SetFloat(fv)
		case reflect.Bool:
			bv, _ := strconv.ParseBool(value)
			v.Field(i).SetBool(bv)
		case reflect.Slice:
			if b := []byte(value); vf.Type == reflect.TypeOf(b) {
				v.Field(i).SetBytes(b)
				continue
			}
			fallthrough
		default:
			return errors.New("field " + vf.Name + ": can't set value to type " + vf.Type.String())
		}
	}

	return nil
}

// 解析 `Tag` 字符串
func parseTag(s string) (tag map[string]string) {
	tag = make(map[string]string)
	setKV := func(k, v string) {
		k = strings.Trim(k, " ")
		v = strings.Trim(v, " ")
		k = strings.Replace(k, "\\", "", -1)
		v = strings.Replace(v, "\\", "", -1)
		if k != "" {
			tag[k] = v
		} else if k == "" && v != "" {
			tag[v] = k
		}
	}

	if s == "-" {
		return
	}

	var k, v string

	i := 0
	for s != "" {
		switch {
		case i == len(s):
			setKV(k, s)
			s = ""
		case s[i] == ':':
			k = s[:i]
			s = s[i+1:]
			i = 0
		case s[i] == ';':
			v = s[:i]
			s = s[i+1:]
			setKV(k, v)
			k, v = "", ""
			i = 0
		case s[i] == '\\' && (s[i+1] == ':' || s[i+1] == ';'):
			i += 2
		default:
			i++
		}
	}

	return
}
