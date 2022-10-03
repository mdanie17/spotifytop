# README

Spotifytop is a tool that enables users to see their top artists and tracks and create a playlist based on those.

To start tool, configure main, e.g.

```go
server := web.Web{
	ServerPort:   "8888",
	State:        "secret",
	RedirectHost: "https://example.org",
	CookieKey:    []byte("secret"),
}
```

The tool uses spotify authentication, and stores the authtoken for the user.
When the user logs out, the token is deleted (the browser might cache the auth process from spotify)
