package web

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
)

func (w *Web) cookieSet(rw http.ResponseWriter, r *http.Request, timelimit string, tracklimit string) {
	cookie := http.Cookie{Name: "settings", Value: timelimit + "," + tracklimit}
	http.SetCookie(rw, &cookie)
}

func cookieGet(rw http.ResponseWriter, r *http.Request, name string) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}

func getSettings(rw http.ResponseWriter, r *http.Request) ([]string, error) {
	cookie, err := cookieGet(rw, r, "settings")
	if err != nil {
		return nil, err
	}
	settings := cookie.Value
	return cookieSettingSplitter(settings), nil
}

func (w *Web) getClient(rw http.ResponseWriter, r *http.Request) spotify.Client {
	token, err := w.Auth.Token(w.State, r)
	if err != nil {
		log.Error().Err(err).Msg("could not get token")
		return w.Client
	}
	w.Client = w.Auth.NewClient(token)
	return w.Client
}

func cookieSettingSplitter(cookiesettings string) []string {
	settings := strings.Split(cookiesettings, ",")
	return settings
}
