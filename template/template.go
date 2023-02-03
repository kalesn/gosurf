package template

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
)

var (
	rw        = new(sync.RWMutex)
	dir       string
	compDir   string
	cacheSize = 10
)

func SetTmplDir(s string) {
	rw.Lock()
	defer rw.Unlock()
	dir = s
}

func SetCompDir(s string) {
	rw.Lock()
	defer rw.Unlock()
	compDir = s

	// to prevent dead lock, use another goroutine
	go RefreshCache()
}

func SetCacheSize(s int) {
	rw.Lock()
	defer rw.Unlock()
	cacheSize = s
}

var (
	components  []string
	tmplCounter = make(map[string]int64)
	tmplCache   = make(map[string]*template.Template)
)

func RefreshCache() {
	rw.Lock()
	defer rw.Unlock()

	// 重置模板计数器及模板缓存
	tmplCounter = make(map[string]int64)
	tmplCache = make(map[string]*template.Template)

	if compDir != "" {
		files, err := filepath.Glob(filepath.Join(dir, compDir, "*.html"))
		if err != nil {
			panic(err)
		}

		components = make([]string, len(files))
		for i, file := range files {
			b, err := ioutil.ReadFile(file)
			if err != nil {
				panic(err)
			}
			s := string(b)
			components[i] = s
		}
	}
}

type Template interface {
	Execute(w io.Writer, data interface{}) error
	ExecuteWithRequest(w http.ResponseWriter, r *http.Request, data map[string]interface{}) error
}

type templateWrapper struct {
	name string
	t    *template.Template
}

func (t *templateWrapper) Execute(w io.Writer, data interface{}) error {
	var buf = new(bytes.Buffer)
	if err := t.t.ExecuteTemplate(buf, t.name, data); err != nil {
		return err
	}
	buf.WriteTo(w)
	return nil
}

func (t *templateWrapper) ExecuteWithRequest(w http.ResponseWriter, r *http.Request, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["request"] = r
	var buf = new(bytes.Buffer)
	if err := t.t.ExecuteTemplate(buf, t.name, data); err != nil {
		return err
	}
	buf.WriteTo(w)
	return nil
}

func Get(name string) Template {
	var tmpl *template.Template

	var exist bool
	withLock(rw.RLocker(), func() {
		tmpl, exist = tmplCache[name]
	})

	if !exist {
		withLock(rw.RLocker(), func() {
			tmpl = newTemplate(name)
			// 当每次生成新模板时，会判断当前的模板缓存的个数是否超过临界值（预设缓存数的两倍）
			// 如果超出了临界值，则调用 `clearCache` 方法进行缓存清理，清理规则是清除使用频率
			// 最低的模板（通过 `tmplCounter` 模板计数器统计），清除后再将当前模板加入模板缓存
			// 此过程会加互斥锁防止并发中的错误操作，整个操作过程会新开一个 `goroutine` 防止
			// 缓存清理影响页面的正常渲染
			go withLock(rw, func() {
				if len(tmplCache) > (cacheSize * 2) {
					clearCache()
				}
				tmplCache[name] = tmpl
			})
		})
	}

	// 边界值溢出检测
	withLock(rw, func() {
		if tmplCounter[name] < (1<<63)-1 {
			tmplCounter[name]++
		}
	})

	return &templateWrapper{filepath.Base(name), tmpl}
}

func newTemplate(name string) (t *template.Template) {
	// 解决了一个 BUG: `NewTemplate` 方法中去除了 `template.New` 新建模板,
	// 因为这里会传入模板的路径作为模板名, 导致模板引擎根据模板名匹配时会出错,
	// 因此去除了通过模板名新建模板的操作, 直接 `ParseFiles`
	// 具体原因待查看模板引擎源码
	t = template.New("").Funcs(funcMap)
	t = template.Must(t.ParseFiles(filepath.Join(dir, name)))
	for _, comp := range components {
		t = template.Must(t.Parse(comp))
	}
	return t
}

type DataMap map[string]interface{}

func RenderWithRequest(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) error {
	return Get(name).ExecuteWithRequest(w, r, data)
}

func Render(w io.Writer, name string, data interface{}) error {
	return Get(name).Execute(w, data)
}

var funcMap = make(template.FuncMap)

func RegisterFunc(name string, fn interface{}) {
	rw.Lock()
	defer rw.Unlock()
	funcMap[name] = fn
}

// `cache` 及 `cacheLists` 是为了对 `map` 进行排序而构建的结构体和排序方法
// `sort.Sort` 方法只能对实现了 `Len`、`Swap`、`Less` 方法的类型进行排序
type cache struct {
	Key   string
	Value int64
}

type cacheLists []cache

func (c cacheLists) Len() int           { return len(c) }
func (c cacheLists) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c cacheLists) Less(i, j int) bool { return c[i].Value < c[j].Value }

// 清理模板缓存，将清理使用频率低的模板，直到模板缓存数小于等于预设的值
func clearCache() {
	// 缓存大小
	n := cacheSize

	// 将模板计数器中的模板名、使用次数生成 `cache` 结构体，并添加到 `cacheLists` 中
	c := make(cacheLists, len(tmplCounter))
	i := 0
	for k, v := range tmplCounter {
		c[i] = cache{k, v}
		i++
	}

	// 对 `cacheLists` 进行从小到大排序
	sort.Sort(c)

	// 若缓存的模板数过多，则从使用频率最低的模板开始删除，直到删除到预设的缓存数
	if l := len(tmplCache); l > n {
		for i, d := 0, l-n; i < d; i++ {
			delete(tmplCache, c[i].Key)
		}
	}

	tmpCache := make(map[string]*template.Template, len(tmplCache))
	for k, v := range tmplCache {
		tmpCache[k] = v
	}

	tmplCache = tmpCache

	// 重置模板计数器
	tmplCounter = make(map[string]int64)
}

func withLock(l sync.Locker, fn func()) {
	l.Lock()
	defer l.Unlock()
	fn()
}
