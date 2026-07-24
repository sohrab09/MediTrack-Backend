package router

import "net/http"

func RegisterRoute(mux *http.ServeMux, apiPrefix, pattern string, handler http.HandlerFunc) {
	var fullPattern string
	spaceIndex := -1
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == ' ' {
			spaceIndex = i
			break
		}
	}

	if spaceIndex != -1 {
		method := pattern[:spaceIndex]
		path := pattern[spaceIndex+1:]
		fullPattern = method + " " + apiPrefix + path
	} else {
		fullPattern = apiPrefix + pattern
	}

	mux.HandleFunc(fullPattern, handler)
}
