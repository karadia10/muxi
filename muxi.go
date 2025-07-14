package muxi

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// Context is passed to every handler and exposes the request/response.
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

// HandlerFunc is the interface for handler function signatures.
type HandlerFunc interface{}

// routeMeta is internal route metadata.
type routeMeta struct {
	Path    string
	Method  string
	Handler HandlerFunc
	InType  reflect.Type
	OutType reflect.Type
}

// App is the core framework struct.
type App struct {
	routes     []*routeMeta
	middleware []func(http.Handler) http.Handler
}

// NewApp returns a new muxi application.
func NewApp() *App {
	return &App{routes: []*routeMeta{}}
}

// RegisterRoute is called by codegen to register annotated handlers.
func (a *App) RegisterRoute(path, method string, h HandlerFunc, in, out reflect.Type) {
	a.routes = append(a.routes, &routeMeta{
		Path: path, Method: method, Handler: h, InType: in, OutType: out,
	})
}

// Use attaches middleware.
func (a *App) Use(mw func(http.Handler) http.Handler) {
	a.middleware = append(a.middleware, mw)
}

// ServeHTTP implements http.Handler.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range a.routes {
		if route.Path == r.URL.Path && route.Method == r.Method {
			ctx := Context{Writer: w, Request: r}
			inVal := reflect.New(route.InType).Interface()

			if route.Method == "POST" || route.Method == "PUT" {
				err := json.NewDecoder(r.Body).Decode(inVal)
				if err != nil {
					http.Error(w, "Invalid JSON", 400)
					return
				}
			} else if route.InType.Kind() == reflect.String {
				val := r.URL.Query().Get("q")
				reflect.ValueOf(inVal).Elem().SetString(val)
			}

			results := reflect.ValueOf(route.Handler).Call([]reflect.Value{
				reflect.ValueOf(ctx), reflect.ValueOf(inVal).Elem(),
			})
			resp := results[0].Interface()
			var err error
			if len(results) > 1 && !results[1].IsNil() {
				err = results[1].Interface().(error)
			}
			if err != nil {
				code := 500
				http.Error(w, err.Error(), code)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}
	http.NotFound(w, r)
}

// Run chains middleware and starts the HTTP server.
func (a *App) Run(addr string) error {
	var handler http.Handler = a
	for i := len(a.middleware) - 1; i >= 0; i-- {
		handler = a.middleware[i](handler)
	}
	return http.ListenAndServe(addr, handler)
}
