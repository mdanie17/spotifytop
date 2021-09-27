package main

import "github.com/mdanie17/spotifytop/web"

func main() {
	server := web.Web{
		ServerPort: "8080",
		State:      "secret",
		Clientkey:  "24a84d4da98541bf89c42f31bea0fd4c",
		Secretkey:  "df69ac40379b4ba78d4688b5b2a540f7",
	}
	server.New()
	server.Run()
}
