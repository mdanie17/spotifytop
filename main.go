package main

import "github.com/mdanie17/spotifytop/web"

func main() {
	server := web.Web{
		ServerPort:   "8888",
		State:        "secret",
		RedirectHost: "https://spotifytop.mdask.dk",
		CookieKey:    []byte("secret"),
	}

	server.New()
	server.Run()
}
