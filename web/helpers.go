package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify/v2"
)

type flashLevel uint8

const (
	flashLevelInfo flashLevel = iota
	flashLevelWarning
	flashLevelDanger
	flashLevelSuccess
)

var (
	ValidTimeLimits = []string{"short_term", "medium_term", "long_term"}
	ErrNoAuthClient = errors.New("no Auth client was found")
)

type TmplData struct {
	Result   interface{}
	Settings Opts
	User     spotify.User
}

type Opts struct {
	Timelimit   string
	Resultlimit int
}

func (fl flashLevel) String() string {
	switch fl {
	case flashLevelInfo:
		return "info"
	case flashLevelWarning:
		return "warning"
	case flashLevelDanger:
		return "danger"
	case flashLevelSuccess:
		return "success"
	}

	log.Error().Msg("could not get flashlevel string")
	return "info"
}

type flashMessage struct {
	Level   flashLevel
	Message string
}

func (w *Web) addFlash(rw http.ResponseWriter, r *http.Request, message flashMessage) {
	session, err := w.Cookies.Get(r, cookieKeyFlashMessage)
	if err != nil {
		log.Error().Err(err).Msg("could not get flash cookie")
		return
	}

	session.AddFlash(message, cookieKeyFlashMessage)
	if err := session.Save(r, rw); err != nil {
		log.Error().Err(err).Msg("could not save session")
		return
	}
}

func (w *Web) getFlash(rw http.ResponseWriter, r *http.Request) []flashMessage {
	session, err := w.Cookies.Get(r, cookieKeyFlashMessage)
	if err != nil {
		log.Error().Err(err).Msg("could not get flash cookie")
		return nil
	}

	flashes := session.Flashes(cookieKeyFlashMessage)

	var flashMessages []flashMessage
	for _, v := range flashes {
		v, ok := v.(flashMessage)
		if !ok {
			continue
		}

		flashMessages = append(flashMessages, v)
	}

	if err := session.Save(r, rw); err != nil {
		log.Error().Err(err).Msg("could not save session")
		return nil
	}

	return flashMessages
}

func (w *Web) cookieSetSettings(rw http.ResponseWriter, r *http.Request, settings Opts) {
	if !checkTimelimit(settings.Timelimit) {
		log.Debug().Interface("timelimit", settings.Timelimit).Msg("unsupported timelimit, using default")
		w.addFlash(rw, r, flashMessage{flashLevelWarning, "You have to select a valid time range"})
		settings.Timelimit = defaultTimeLimit
	}

	cookie := http.Cookie{Name: "settings", Value: settings.Timelimit + "," + strconv.Itoa(settings.Resultlimit)}
	http.SetCookie(rw, &cookie)
}

func cookieGet(rw http.ResponseWriter, r *http.Request, name string) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

func (w *Web) cookieGetSettings(rw http.ResponseWriter, r *http.Request) Opts {
	cookie, err := cookieGet(rw, r, "settings")
	settings := Opts{defaultTimeLimit, defaultResultLimit}

	if err != nil {
		if err == http.ErrNoCookie {
			w.cookieSetSettings(rw, r, settings)
			return settings
		}

		log.Error().Err(err).Msg("could not get settings, using defaults")
		return settings
	}

	cookiesettings := cookieSettingSplitter(cookie.Value)
	settings.Timelimit = cookiesettings[0]
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

	if w.Auth == nil {
		http.Redirect(rw, r, "/", http.StatusFound)
		return ErrNoAuthClient
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

func redirectReferer(rw http.ResponseWriter, r *http.Request) {
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}

	http.Redirect(rw, r, ref, http.StatusSeeOther)
}
