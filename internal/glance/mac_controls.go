package glance

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/glanceapp/glance/pkg/sysinfo"
)

type macSpotifyStatus struct {
	State   string `json:"state"`
	Track   string `json:"track"`
	Artist  string `json:"artist"`
	Artwork string `json:"artwork"`
}

type macStats struct {
	CPU uint8 `json:"cpu"`
	RAM uint8 `json:"ram"`
}

type macStatusResponse struct {
	Volume  int              `json:"volume"`
	Muted   bool             `json:"muted"`
	Spotify macSpotifyStatus `json:"spotify"`
	Stats   macStats         `json:"stats"`
}

func runAppleScript(script string) string {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getMacStatus() macStatusResponse {
	script := `
set curVol to output volume of (get volume settings)
set isMuted to output muted of (get volume settings)
set spotState to "offline"
set spotTrack to ""
set spotArtist to ""
set spotArt to ""
if application "Spotify" is running then
    tell application "Spotify"
        set spotState to player state as string
        try
            set spotTrack to name of current track
            set spotArtist to artist of current track
            set spotArt to artwork url of current track
        end try
    end tell
end if
return (curVol as string) & ":::" & (isMuted as string) & ":::" & spotState & ":::" & spotTrack & ":::" & spotArtist & ":::" & spotArt
`
	raw := runAppleScript(script)
	parts := strings.Split(raw, ":::")
	res := macStatusResponse{
		Volume: 0,
		Muted:  false,
		Spotify: macSpotifyStatus{
			State: "offline",
		},
	}
	if len(parts) >= 6 {
		if v, err := strconv.Atoi(parts[0]); err == nil {
			res.Volume = v
		}
		res.Muted = (strings.ToLower(parts[1]) == "true")
		res.Spotify.State = parts[2]
		res.Spotify.Track = parts[3]
		res.Spotify.Artist = parts[4]
		res.Spotify.Artwork = parts[5]
	}

	info, _ := sysinfo.Collect(nil)
	if info != nil {
		res.Stats.CPU = info.CPU.Load1Percent
		res.Stats.RAM = info.Memory.UsedPercent
	}

	return res
}

func handleMacControls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path := r.URL.Path
	q := r.URL.Query()

	if strings.HasSuffix(path, "/volume") {
		action := q.Get("action")
		setVal := q.Get("set")
		if setVal != "" {
			if v, err := strconv.Atoi(setVal); err == nil {
				runAppleScript("set volume output volume " + strconv.Itoa(v))
			}
		} else if action == "up" {
			runAppleScript("set volume output volume ((output volume of (get volume settings)) + 6)")
		} else if action == "down" {
			runAppleScript("set volume output volume ((output volume of (get volume settings)) - 6)")
		} else if action == "mute" {
			runAppleScript("set volume output muted (not (output muted of (get volume settings)))")
		}
	} else if strings.HasSuffix(path, "/spotify") {
		action := q.Get("action")
		if action == "playpause" {
			runAppleScript(`tell application "Spotify" to playpause`)
		} else if action == "next" {
			runAppleScript(`tell application "Spotify" to next track`)
		} else if action == "prev" {
			runAppleScript(`tell application "Spotify" to previous track`)
		} else if action == "launch" {
			runAppleScript(`tell application "Spotify" to activate`)
		}
	}

	data := getMacStatus()
	_ = json.NewEncoder(w).Encode(data)
}
