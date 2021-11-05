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
	LoggedIn bool
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

func (o Opts) TimeLimitFormatter() string {
	switch o.Timelimit {
	case "short_term":
		return "Last month"
	case "medium_term":
		return "Last 6 months"
	case "long_term":
		return "All time"
	default:
		return "Undefined timelimit"
	}
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

	cookie := http.Cookie{Name: "settings", Value: settings.Timelimit + "," + strconv.Itoa(settings.Resultlimit), MaxAge: 3600}
	http.SetCookie(rw, &cookie)
}

func cookieSetState(rw http.ResponseWriter, r *http.Request, state string) {
	http.SetCookie(rw, &http.Cookie{Name: "state", Value: state, MaxAge: 3600})
}

func (w *Web) cookieGetState(rw http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := cookieGet(rw, r, "state")
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
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

func (w *Web) createClient(rw http.ResponseWriter, r *http.Request, state string) error {
	if w.Auth == nil {
		return ErrNoAuthClient
	}

	if _, ok := w.Clients[state]; ok {
		return nil
	}

	token, err := w.Auth.Token(r.Context(), state, r)
	if err != nil {
		log.Error().Err(err).Msg("could not get token")
		return err
	}

	w.Clients[state] = spotify.New(w.Auth.Client(r.Context(), token))
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

func (w *Web) deleteCookie(rw http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(rw, &http.Cookie{Name: name, Value: ""})
}

func getTrackIDs(tracks []spotify.FullTrack) []spotify.ID {
	var ids []spotify.ID
	for _, track := range tracks {
		ids = append(ids, track.ID)
	}

	return ids
}
