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
		// THIS SHIT IS TEMPORARY REMOVVVVVVVVVVVVVVVVVVVEEEEEEEEEEEEEEEEEEEEEEEEEEE
		log.Printf("Spotify currently-playing rate limited, Retry-After: %s", resp.Header.Get("Retry-After"))
		dummy := PlayResponse{
			IsPlaying:  false,
			Timestamp:  0,
			ProgressMs: 97000,
			DurationMs: 210000,
			Track:      "After LIKE",
			Artist:     "IVE",
			Album:      "After LIKE",
			AlbumArt:   "data:image/png;base64,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",
			URL:        "https://sukaseven.com",
		}
		util.WriteJSON(w, http.StatusTooManyRequests, dummy)
		return // THIS SHIT IS TEMPORARY REMOVVVVVVVVVVVVVVVVVVVEEEEEEEEEEEEEEEEEEEEEEEEEEE
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
