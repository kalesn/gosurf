package form

import (
	"errors"
	"gosurf/util"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Scanner interface {
	Scan(v interface{}) error
}

// `Clean` 方法传入 `form` 结构体指针以及由 `net.Value`，
// 随后调用相应的方法对表单进行清洗，出现错误则返回错误。
func Clean(formPtr interface{}, postForm url.Values) (error, error) {
	v := reflect.ValueOf(formPtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, errors.New("expect ptr value")
	}

	formType := reflect.TypeOf(formPtr).Elem()
	formValue := reflect.Indirect(v)

	if formValue.Kind() != reflect.Struct {
		return nil, errors.New("expect struct value")
	}

	// 循环清洗每一个字段
	for i := 0; i < formType.NumField(); i++ {
		sf := formType.Field(i)

		// skip anonymous field
		if sf.Anonymous {
			continue
		}

		fv := formValue.Field(i)
		tag := parseTag(sf.Tag.Get("form"))

		// 如果 `Tag` 中未提供 `name` 信息，则使用字段的下划线命名法名称
		name, ok := tag["name"]
		if !ok {
			name = util.SnakeCase(sf.Name)
		}

		// 从这里开始正式清洗表单信息，先获取由页面传来的表单内容，
		// 随后调用相应的清洗方法进行清洗，遇到错误直接返回错误，
		// 最后将清洗后的值赋给表单字段。
		postv := postForm[name]
		if len(postv) == 0 {
			continue
		}

		if meth, ok := tag["method"]; !ok || meth == "-" {
			if scanner, ok := fv.Addr().Interface().(Scanner); ok {
				scanner.Scan(postv[0])
			} else if fv.Kind() == reflect.String {
				fv.SetString(postv[0])
			}
			continue
		}

		// 如果 `Tag` 中未提供 `meth` 信息，则跳过不清洗
		meth := formValue.MethodByName(tag["method"])
		if !meth.IsValid() {
			return nil, errors.New("error clean method")
		}

		cleanValue := meth.Call([]reflect.Value{reflect.ValueOf(postv[0])})

		switch len(cleanValue) {
		case 1:
		case 2:
			errv := cleanValue[1]
			if errv.Type() != reflect.TypeOf(new(error)).Elem() {
				return nil, errors.New("second return value of clean method should be error")
			}
			if !errv.IsNil() {
				return errv.Interface().(error), nil
			}
		default:
			return nil, errors.New("clean method should return 1 or 2 values")
		}

		// cleaned value
		cv := cleanValue[0]

		// try to scan in value
		if scanner, ok := fv.Addr().Interface().(Scanner); ok {
			err := scanner.Scan(cv.Interface())
			if err != nil {
				return nil, err
			}
			continue
		}

		if fvT, cvT := sf.Type, cv.Type(); fvT != cvT {
			if cvT.ConvertibleTo(fvT) {
				fv.Set(cv.Convert(fvT))
			} else {
				return nil, errors.New("can not assign type " + cvT.String() + " to type " + fvT.String())
			}
		} else {
			fv.Set(cv)
		}
	}

	return nil, nil
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

const (
	skipRoot  = "root"
	skipAdmin = "admin"
)

type Validator struct{}

var rePureNum = regexp.MustCompile(`^\d+$`)

func (Validator) ValidateUsername(s string) (username string, err error) {
	if s == skipRoot || s == skipAdmin {
		return s, nil
	}
	if l := len(s); l < 6 || l > 20 {
		err = errors.New("username must be 6 ~ 20 charactor")
		return
	}
	if reEmail.MatchString(s) {
		err = errors.New("username should not be an email address")
		return
	}
	if rePhoneNum.MatchString(s) {
		err = errors.New("username should not be a phone number")
		return
	}
	if rePureNum.MatchString(s) {
		err = errors.New("username should not be pure numbers")
		return
	}
	username = s
	return
}

var reEmail = regexp.MustCompile(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`)

func (Validator) ValidateEmail(s string) (email string, err error) {
	if !reEmail.MatchString(s) {
		err = errors.New("not a valid email address")
		return
	}
	email = s
	return
}

var (
	reNum   = regexp.MustCompile(`\d`)
	reUpper = regexp.MustCompile(`[A-Z]`)
	reLower = regexp.MustCompile(`[a-z]`)
)

func (Validator) ValidatePassword(s string) (password string, err error) {
	if len(s) < 6 {
		err = errors.New("password must be at least 6 charactor")
		return
	}
	if !reNum.MatchString(s) {
		err = errors.New("password must contain number")
		return
	}
	if !reUpper.MatchString(s) {
		err = errors.New("password must contain upper case letter")
		return
	}
	if !reLower.MatchString(s) {
		err = errors.New("password must contain lower case letter")
		return
	}
	password = s
	return
}

var rePhoneNum = regexp.MustCompile(`^(?:(?:13[0-9])|(?:14[5,7])|(?:15[0-3,5-9])|(?:17[03,5-8])|(?:18[0-9])|166|198|199|147)\d{8}$`)

func (Validator) ValidatePhoneNumber(s string) (phone string, err error) {
	if !rePhoneNum.MatchString(s) {
		err = errors.New("not a valid phone number")
		return
	}
	phone = s
	return
}

func (Validator) ParseInt(s string) (i int64, err error) {
	return strconv.ParseInt(s, 10, 64)
}
