package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify/v2"
)

var (
	ErrConvertResultFault = errors.New("could not convert result to int")
	ValidTimeLimits       = []string{"short_term", "medium_term", "long_term"}
)

func (w *Web) cookieSet(rw http.ResponseWriter, r *http.Request, settings Opts) {
	cookie := http.Cookie{Name: "settings", Value: settings.Timelimit + "," + fmt.Sprint(settings.Resultlimit)}
	http.SetCookie(rw, &cookie)
}

func cookieGet(rw http.ResponseWriter, r *http.Request, name string) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}

//TODO eventually display errors as flash message
func (w *Web) getSettings(rw http.ResponseWriter, r *http.Request) Opts {
	cookie, err := cookieGet(rw, r, "settings")
	settings := Opts{defaultTimeLimit, defaultResultLimit}

	if err != nil {
		if err == http.ErrNoCookie {
			w.cookieSet(rw, r, settings)
			return settings
		}

		log.Error().Err(err).Msg("could not get settings, using defaults")
		return settings
	}

	cookiesettings := cookieSettingSplitter(cookie.Value)
	settings.Timelimit = cookiesettings[0]
	if !checkTimelimit(settings.Timelimit) {
		log.Error().Err(err).Msg("unsupported timelimit, using default")
		settings.Timelimit = defaultTimeLimit
	}

	settings.Resultlimit, err = strconv.Atoi(cookiesettings[1])
	if err != nil {
		log.Error().Err(err).Msgf("could not convert %s to int, using default result limit", cookiesettings[1])
		settings.Resultlimit = defaultResultLimit
		return settings
	}

	return settings
}

func (w *Web) getClient(rw http.ResponseWriter, r *http.Request) error {
	if w.Client != nil {
		return nil
	}

	token, err := w.Auth.Token(r.Context(), w.State, r)
	if err != nil {
		log.Error().Err(err).Msg("could not get token")
		return err
	}

	w.Client = spotify.New(w.Auth.Client(r.Context(), token))
	return nil
}

func cookieSettingSplitter(cookiesettings string) []string {
	settings := strings.Split(cookiesettings, ",")
	return settings
}

func checkTimelimit(timelimit string) bool {
	for _, validlimit := range ValidTimeLimits {
		if timelimit == validlimit {
			return true
		}
	}
	return false
}
