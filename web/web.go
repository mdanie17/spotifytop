package web

import (
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
)

var (
	defaultTimeLimit   = "medium"
	defaultResultLimit = 20
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

type Web struct {
	Router *mux.Router

	ServerPort string

	Templates map[string]*template.Template

	State     string
	Auth      spotify.Authenticator
	Clientkey string
	Secretkey string

	Client spotify.Client
}

func (w *Web) New() {
	if w.Router == nil {
		w.Router = mux.NewRouter()
		w.Routes(w.Router)
	}

	if w.ServerPort == "" {
		w.ServerPort = "8080"
	}

	if w.Templates == nil {
		w.Templates = make(map[string]*template.Template)

		w.parseTemplate("topartists", "")
		w.parseTemplate("frontpage", "")
		w.parseTemplate("toptracks", "")
	}

	if w.State == "" {
		log.Fatal().Msg("you have to set a state string")
		return
	}

	if w.Clientkey == "" {
		log.Fatal().Msg("you have to set a client key")
	}

	if w.Secretkey == "" {
		log.Fatal().Msg("you have to set a secret key")
	}
}

func (w *Web) Routes(r *mux.Router) {
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./web/templates/css"))))

	r.HandleFunc("/", w.handleFrontPage).Methods("GET")
	r.HandleFunc("/topartistsauth", w.handleAuthenticateArtists)
	r.HandleFunc("/topartists", w.handleTopArtists)
	r.HandleFunc("/toptracksauth", w.handleAuthenticateTracks)
	r.HandleFunc("/toptracks", w.handleTopTracks)
	r.HandleFunc("/form", w.handleForm)
}

func (w *Web) Run() {
	log.Info().Msgf("Starting server on port %s", w.ServerPort)
	if err := http.ListenAndServe("localhost:"+w.ServerPort, w.Router); err != nil {
		log.Fatal().Err(err).Msg("failed to start webserver")
	}
}

func (w *Web) handleFrontPage(rw http.ResponseWriter, r *http.Request) {
	var timeframelimit string
	var resultlimit int

	settings, err := getSettings(rw, r)
	if err != nil {
		timeframelimit = defaultTimeLimit
		resultlimit = defaultResultLimit
	} else {
		timeframelimit = settings[0]
		resultlimit, err = strconv.Atoi(settings[1])
		if err != nil {
			log.Error().Err(err).Msgf("could not convert result limit to int: %s", settings[1])
			resultlimit = defaultResultLimit
		}
	}

	w.templateExec(rw, r, "frontpage", TmplData{Settings: Opts{timeframelimit, resultlimit}})
}

func (w *Web) handleTopArtists(rw http.ResponseWriter, r *http.Request) {

	client := w.getClient(rw, r)
	user, err := client.CurrentUser()
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
	}

	var timeframelimit string
	var artistlimit int
	settings, err := getSettings(rw, r)
	if err != nil {
		timeframelimit = defaultTimeLimit
		artistlimit = defaultResultLimit
	} else {
		timeframelimit = settings[0]
		artistlimit, err = strconv.Atoi(settings[1])
		if err != nil {
			log.Error().Err(err).Msgf("could not convert artistlimit setting to int: %s", settings[1])
			artistlimit = defaultResultLimit
		}
	}

	if timeframelimit == "" {
		timeframelimit = defaultTimeLimit
	}

	topartists, err := client.CurrentUsersTopArtistsOpt(&spotify.Options{Timerange: &timeframelimit, Limit: &artistlimit})
	if err != nil {
		log.Error().Err(err).Msg("could not get current user top artists")
	}

	Data := TmplData{
		Result:   topartists.Artists,
		Settings: Opts{timeframelimit, artistlimit},
		User:     user.User,
	}
	w.templateExec(rw, r, "topartists", Data)
}

func (w *Web) handleTopTracks(rw http.ResponseWriter, r *http.Request) {
	client := w.getClient(rw, r)
	user, err := client.CurrentUser()
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
	}

	var timeframelimit string
	var tracklimit int
	settings, err := getSettings(rw, r)
	if err != nil {
		timeframelimit = defaultTimeLimit
		tracklimit = defaultResultLimit
	} else {
		timeframelimit = settings[0]
		tracklimit, err = strconv.Atoi(settings[1])
		if err != nil {
			log.Error().Err(err).Msg("could not convert artistlimit setting to int")
			tracklimit = defaultResultLimit
		}
	}

	if timeframelimit == "" {
		timeframelimit = defaultTimeLimit
	}
	toptracks, err := client.CurrentUsersTopTracksOpt(&spotify.Options{Timerange: &timeframelimit, Limit: &tracklimit})
	if err != nil {
		log.Error().Err(err).Msg("could not get current user top tracks")
	}

	Data := TmplData{
		Result:   toptracks.Tracks,
		Settings: Opts{timeframelimit, tracklimit},
		User:     user.User,
	}
	w.templateExec(rw, r, "toptracks", Data)
}

func (w *Web) handleAuthenticateArtists(rw http.ResponseWriter, r *http.Request) {
	w.Auth = spotify.NewAuthenticator("http://localhost:8080/topartists", spotify.ScopeUserTopRead, spotify.ScopeUserReadPrivate)
	w.Auth.SetAuthInfo(w.Clientkey, w.Secretkey)
	w.getClient(rw, r)
	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}

func (w *Web) handleAuthenticateTracks(rw http.ResponseWriter, r *http.Request) {
	w.Auth = spotify.NewAuthenticator("http://localhost:8080/toptracks", spotify.ScopeUserTopRead, spotify.ScopeUserReadPrivate)
	w.Auth.SetAuthInfo(w.Clientkey, w.Secretkey)
	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}

func (w *Web) handleForm(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("could not parse settings form")
	}

	timelimit, artistlimit := r.FormValue("timecheck"), r.FormValue("limit")

	w.cookieSet(rw, r, timelimit, artistlimit)
	http.Redirect(rw, r, r.Referer(), http.StatusFound)
}
