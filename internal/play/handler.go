package play

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/suka712/api.sukaseven.com/util"
)

type SpotifyImage struct {
	URL string `json:"url"`
}

type SpotifyArtist struct {
	Name string `json:"name"`
}

type SpotifyAlbum struct {
	Name   string         `json:"name"`
	Images []SpotifyImage `json:"images"`
}

type SpotifyTrack struct {
	Name         string          `json:"name"`
	DurationMs   int64           `json:"duration_ms"`
	Artists      []SpotifyArtist `json:"artists"`
	Album        SpotifyAlbum    `json:"album"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

type CurrentlyPlayingResponse struct {
	CurrentlyPlayingType string       `json:"currently_playing_type"`
	IsPlaying            bool         `json:"is_playing"`
	Timestamp            int64        `json:"timestamp"`
	ProgressMs           int64        `json:"progress_ms"`
	Item                 SpotifyTrack `json:"item"`
}

type RecentlyPlayedResponse struct {
	Items []struct {
		Track SpotifyTrack `json:"track"`
	} `json:"items"`
}

type PlayResponse struct {
	IsPlaying  bool   `json:"is_playing"`
	Timestamp  int64  `json:"timestamp"`
	ProgressMs int64  `json:"progress_ms"`
	DurationMs int64  `json:"duration_ms"`
	Track      string `json:"track"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	AlbumArt   string `json:"album_art"`
	URL        string `json:"url"`
}

func trackToResponse(track SpotifyTrack, isPlaying bool, timestamp int64, progressMs int64) PlayResponse {
	albumArt := ""
	if len(track.Album.Images) > 0 {
		albumArt = track.Album.Images[0].URL
	}

	return PlayResponse{
		IsPlaying:  isPlaying,
		Timestamp:  timestamp,
		ProgressMs: progressMs,
		DurationMs: track.DurationMs,
		Track:      track.Name,
		Artist:     track.Artists[0].Name,
		Album:      track.Album.Name,
		AlbumArt:   albumArt,
		URL:        track.ExternalURLs.Spotify,
	}
}

func Play(w http.ResponseWriter, r *http.Request) {
	resp, err := spotifyGet("https://api.spotify.com/v1/me/player/currently-playing")
	if err != nil {
		log.Print("Error fetching currently playing:", err)
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var current CurrentlyPlayingResponse
		json.NewDecoder(resp.Body).Decode(&current)

		if current.CurrentlyPlayingType == "track" {
			util.WriteJSON(w, http.StatusOK, trackToResponse(current.Item, current.IsPlaying, current.Timestamp, current.ProgressMs))
			return
		}
	} else if resp.StatusCode == 429 {
		log.Printf("Spotify currently-playing rate limited, Retry-After: %s", resp.Header.Get("Retry-After"))
	} else if resp.StatusCode != 204 {
		log.Printf("Spotify currently-playing returned status %d", resp.StatusCode)
	}

	recentResp, err := spotifyGet("https://api.spotify.com/v1/me/player/recently-played?limit=1")
	if err != nil {
		log.Print("Error fetching recently played:", err)
		util.WriteJSON(w, http.StatusInternalServerError, util.ErrorResponse{Error: "Something went wrong"})
		return
	}
	defer recentResp.Body.Close()

	if recentResp.StatusCode == 429 {
		log.Printf("Spotify recently-played rate limited, Retry-After: %s", recentResp.Header.Get("Retry-After"))
		util.WriteJSON(w, http.StatusOK, PlayResponse{})
		return
	} else if recentResp.StatusCode != 200 {
		log.Printf("Spotify recently-played returned status %d", recentResp.StatusCode)
		util.WriteJSON(w, http.StatusOK, PlayResponse{})
		return
	}

	var recent RecentlyPlayedResponse
	json.NewDecoder(recentResp.Body).Decode(&recent)

	if len(recent.Items) == 0 {
		util.WriteJSON(w, http.StatusOK, PlayResponse{})
		return
	}

	util.WriteJSON(w, http.StatusOK, trackToResponse(recent.Items[0].Track, false, 0, 0))
}
