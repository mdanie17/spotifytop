package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aidarkhanov/nanoid"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func (w *Web) handleTopArtists(rw http.ResponseWriter, r *http.Request) {
	state, err := w.cookieGetState(rw, r)
	if err != nil {
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "You have to log in first"})
		redirectReferer(rw, r)
		return
	}

	client := w.Clients[state]

	user, err := client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	settings := w.cookieGetSettings(rw, r)

	topartists, err := client.CurrentUsersTopArtists(
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
		LoggedIn: true,
	}

	w.templateExec(rw, r, "topartists", Data)
}

func (w *Web) handleTopTracks(rw http.ResponseWriter, r *http.Request) {
	state, err := w.cookieGetState(rw, r)
	if err != nil {
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "You have to log in first"})
		redirectReferer(rw, r)
		return
	}

	client := w.Clients[state]

	user, err := client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	settings := w.cookieGetSettings(rw, r)

	toptracks, err := client.CurrentUsersTopTracks(
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
		LoggedIn: true,
	}

	w.templateExec(rw, r, "toptracks", Data)
}

func (w *Web) handleAuth(rw http.ResponseWriter, r *http.Request) {
	state := nanoid.New()
	cookieSetState(rw, r, state)
	w.Auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(fmt.Sprintf("http://%s/authenticated", w.RedirectHost)),
		spotifyauth.WithScopes(spotifyauth.ScopeUserTopRead, spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistModifyPrivate),
		spotifyauth.WithClientID(w.Clientkey),
		spotifyauth.WithClientSecret(w.Secretkey),
	)

	http.Redirect(rw, r, w.Auth.AuthURL(state), http.StatusFound)
}

//TODO(mdask) Maybe look for if playlist already exists, and overwrite it??
func (w *Web) handleCreatePlaylist(rw http.ResponseWriter, r *http.Request) {
	state, err := w.cookieGetState(rw, r)
	if err != nil {
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "You have to log in first"})
		redirectReferer(rw, r)
		return
	}

	client := w.Clients[state]

	user, err := client.CurrentUser(r.Context())
	if err != nil {
		log.Error().Err(err).Msgf("could not get user")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	settings := w.cookieGetSettings(rw, r)

	playlistname := fmt.Sprintf("%s Top %d tracks", user.DisplayName, settings.Resultlimit)
	creationyear, creationmonth, creationday := time.Now().Date()
	playlistdesc := fmt.Sprintf("%s Top %d tracks - %s | Created %v %v %v", user.DisplayName, settings.Resultlimit, settings.TimeLimitFormatter(), creationyear, creationmonth, creationday)

	playlist, err := client.CreatePlaylistForUser(r.Context(), user.ID, playlistname, playlistdesc, false, false)
	if err != nil {
		log.Error().Err(err).Msg("could not create playlist")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	toptracks, err := client.CurrentUsersTopTracks(
		r.Context(),
		spotify.Limit(settings.Resultlimit),
		spotify.Timerange(spotify.Range(settings.Timelimit)),
	)
	if err != nil {
		log.Error().Err(err).Msg("could not create playlist")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	_, err = client.AddTracksToPlaylist(r.Context(), playlist.ID, getTrackIDs(toptracks.Tracks)...)
	if err != nil {
		log.Error().Err(err).Msg("could not create playlist")
		w.addFlash(rw, r, flashMessage{flashLevelDanger, "Could not communicate with Spotify - Try clearing cache and trying again"})
		redirectReferer(rw, r)
		return
	}

	w.addFlash(rw, r, flashMessage{flashLevelSuccess, "Succesfully created playlist"})
	redirectReferer(rw, r)
}
