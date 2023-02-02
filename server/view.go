package server

import (
	"net/http"
)

type View struct {
	Name    string
	Desc    string
	Get     func(w http.ResponseWriter, r *http.Request)
	Head    func(w http.ResponseWriter, r *http.Request)
	Post    func(w http.ResponseWriter, r *http.Request)
	Options func(w http.ResponseWriter, r *http.Request)
	Put     func(w http.ResponseWriter, r *http.Request)
	Delete  func(w http.ResponseWriter, r *http.Request)
	Trace   func(w http.ResponseWriter, r *http.Request)
	Connect func(w http.ResponseWriter, r *http.Request)
	Patch   func(w http.ResponseWriter, r *http.Request)

	// 405 method not allowed func
	func405 func(w http.ResponseWriter, r *http.Request)
}

func (v View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if v.Get != nil {
			v.Get(w, r)
			return
		}
	case http.MethodHead:
		if v.Head != nil {
			v.Head(w, r)
			return
		}
	case http.MethodPost:
		if v.Post != nil {
			v.Post(w, r)
			return
		}
	case http.MethodOptions:
		if v.Options != nil {
			v.Options(w, r)
			return
		}
	case http.MethodPut:
		if v.Put != nil {
			v.Put(w, r)
			return
		}
	case http.MethodDelete:
		if v.Delete != nil {
			v.Delete(w, r)
			return
		}
	case http.MethodTrace:
		if v.Trace != nil {
			v.Trace(w, r)
			return
		}
	case http.MethodConnect:
		if v.Connect != nil {
			v.Connect(w, r)
			return
		}
	case http.MethodPatch:
		if v.Patch != nil {
			v.Patch(w, r)
			return
		}
	default:
		if v.func405 != nil {
			v.func405(w, r)
			return
		}
	}
}
