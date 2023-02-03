package gosurf_demo

import (
	"gosurf/server"
	"gosurf/template"
	"gosurf/util"
	"net/http"
	"net/http/pprof"
)

func init() {
	server.Handle(`/api/logout/`, logoutAPI)

	//刷新模板缓存
	server.Handle(`/api/refresh/`, refreshTmpl)

	//pprof debug api
	server.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	server.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	server.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	server.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	server.Handle("/debug/pprof/.*", http.HandlerFunc(pprof.Index))
}

var (
	logoutAPI = server.View{
		Name: "logout_api",
		Desc: "登出接口",
		Post: func(w http.ResponseWriter, r *http.Request) {
			//ctx := r.Context()

			//s, exist := common.RequestSession(r)
			//if exist {
			//	common.SessionManager.Remove(ctx, s)
			//}

			cookie := http.Cookie{
				//Name:   common.SessionKey,
				Name:   "Test",
				Path:   "/",
				MaxAge: -1,
			}
			http.SetCookie(w, &cookie)

			return
		},
	}

	refreshTmpl = server.View{
		Name: "refresh_template",
		Desc: "刷新模板缓存接口",
		Get: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var dat = make(util.Json)
			defer dat.ResponseTo(ctx, w)

			template.RefreshCache()
			//handler.ServeFile("./static/img/fine.png")

			dat.SetCode(0)
		},
	}
)
