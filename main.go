package main

import "github.com/mdanie17/spotifytop/web"

func main() {
	server := web.Web{
		ServerPort: "8080",
		State:      "secret",
	}

	server.New()
	server.Run()
}
