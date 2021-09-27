package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	defaultTimeLimit   = "medium_term"
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
	Auth      *spotifyauth.Authenticator
	Clientkey string
	Secretkey string

	Client *spotify.Client
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
		if w.Clientkey = os.Getenv("SPOTIFY_ID"); w.Clientkey == "" {
			fmt.Println(os.Getenv("SPOTIFY_ID"))

			log.Fatal().Msg("you have to set a client key")
		}
	}

	if w.Secretkey == "" {
		if w.Secretkey = os.Getenv("SPOTIFY_SECRET"); w.Secretkey == "" {
			log.Fatal().Msg("you have to set a secret key")
		}
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
	settings := w.getSettings(rw, r)

	w.templateExec(rw, r, "frontpage", TmplData{Settings: settings})
}

func (w *Web) handleTopArtists(rw http.ResponseWriter, r *http.Request) {
	if err := w.getClient(rw, r); err != nil {
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		log.Error().Err(err).Msg("could not get client")
		return
	}

	user, err := w.Client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
	}

	settings := w.getSettings(rw, r)

	topartists, err := w.Client.CurrentUsersTopArtists(
		r.Context(),
		spotify.Limit(settings.Resultlimit),
		spotify.Timerange(spotify.Range(settings.Timelimit)),
	)

	if err != nil {
		log.Error().Err(err).Msg("could not get current user top artists")
	}

	Data := TmplData{
		Result:   topartists.Artists,
		Settings: Opts{settings.Timelimit, settings.Resultlimit},
		User:     user.User,
	}
	w.templateExec(rw, r, "topartists", Data)
}

func (w *Web) handleTopTracks(rw http.ResponseWriter, r *http.Request) {
	if err := w.getClient(rw, r); err != nil {
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		log.Error().Err(err).Msg("could not get client")
		return
	}

	user, err := w.Client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
	}

	settings := w.getSettings(rw, r)

	toptracks, err := w.Client.CurrentUsersTopTracks(
		r.Context(),
		spotify.Limit(settings.Resultlimit),
		spotify.Timerange(spotify.Range(settings.Timelimit)),
	)

	if err != nil {
		log.Error().Err(err).Msg("could not get current user top tracks")
		return
	}

	Data := TmplData{
		Result:   toptracks.Tracks,
		Settings: Opts{settings.Timelimit, settings.Resultlimit},
		User:     user.User,
	}
	w.templateExec(rw, r, "toptracks", Data)
}

func (w *Web) handleAuthenticateArtists(rw http.ResponseWriter, r *http.Request) {
	w.Auth = spotifyauth.New(
		spotifyauth.WithRedirectURL("http://localhost:8080/topartists"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserTopRead, spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(w.Clientkey),
		spotifyauth.WithClientSecret(w.Secretkey),
	)
	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}

func (w *Web) handleAuthenticateTracks(rw http.ResponseWriter, r *http.Request) {
	w.Auth = spotifyauth.New(
		spotifyauth.WithRedirectURL("http://localhost:8080/toptracks"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserTopRead, spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(w.Clientkey),
		spotifyauth.WithClientSecret(w.Secretkey),
	)
	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}

func (w *Web) handleForm(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("could not parse settings form")
	}
	timelimit := r.FormValue("timecheck")
	resultlimit := r.FormValue("limit")
	resultlimitint, err := strconv.Atoi(resultlimit)
	if err != nil {
		log.Error().Err(err).Msg("could not convert result limit to int, using default")
		resultlimitint = defaultResultLimit
	}

	w.cookieSet(rw, r, Opts{timelimit, resultlimitint})
	http.Redirect(rw, r, r.Referer(), http.StatusFound)
}
