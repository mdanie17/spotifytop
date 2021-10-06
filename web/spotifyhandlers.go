package web

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func (w *Web) handleTopArtists(rw http.ResponseWriter, r *http.Request) {
	if err := w.getClient(rw, r); err != nil {
		log.Error().Err(err).Msg("could not get client")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	user, err := w.Client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	settings := w.cookieGetSettings(rw, r)

	topartists, err := w.Client.CurrentUsersTopArtists(
		r.Context(),
		spotify.Limit(settings.Resultlimit),
		spotify.Timerange(spotify.Range(settings.Timelimit)),
	)

	if err != nil {
		log.Error().Err(err).Msg("could not get current user top artists")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
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
		log.Error().Err(err).Msg("could not get client")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	user, err := w.Client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	settings := w.cookieGetSettings(rw, r)

	toptracks, err := w.Client.CurrentUsersTopTracks(
		r.Context(),
		spotify.Limit(settings.Resultlimit),
		spotify.Timerange(spotify.Range(settings.Timelimit)),
	)

	if err != nil {
		log.Error().Err(err).Msg("could not get current user top tracks")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
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
		spotifyauth.WithRedirectURL(fmt.Sprintf("http://%s:%s/topartists", w.ServerHostName, w.ServerPort)),
		spotifyauth.WithScopes(spotifyauth.ScopeUserTopRead, spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(w.Clientkey),
		spotifyauth.WithClientSecret(w.Secretkey),
	)

	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}

func (w *Web) handleAuthenticateTracks(rw http.ResponseWriter, r *http.Request) {
	w.Auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(fmt.Sprintf("http://%s:%s/toptracks", w.ServerHostName, w.ServerPort)),
		spotifyauth.WithScopes(spotifyauth.ScopeUserTopRead, spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(w.Clientkey),
		spotifyauth.WithClientSecret(w.Secretkey),
	)

	http.Redirect(rw, r, w.Auth.AuthURL(w.State), http.StatusFound)
}
