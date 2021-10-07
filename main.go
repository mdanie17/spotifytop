package main

import "github.com/mdanie17/spotifytop/web"

func main() {
	server := web.Web{
		ServerPort:   "8080",
		State:        "secret",
		RedirectHost: "localhost:8080",
		CookieKey:    []byte("secret"),
	}

	server.New()
	server.Run()
}
