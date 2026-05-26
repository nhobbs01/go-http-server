package main

type HandlerFunc func(req HTTPRequest) string

// Capitalized to make it globally accesible

var Routes = make(map[string]HandlerFunc)

func RegisterRoute(path string, handler HandlerFunc) {
	Routes[path] = handler
}
