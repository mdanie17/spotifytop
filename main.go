package main

import "gospotify/web"

func main() {
	server := web.Web{
		ServerPort: "8080",
		State:      "secret",
		Clientkey:  "***REMOVED***",
		Secretkey:  "***REMOVED***",
	}
	server.New()
	server.Run()
}
