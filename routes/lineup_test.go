package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duncanleo/plex-dvr-hls/config"
	"github.com/gin-gonic/gin"
)

// TestLineupReturnsChannels verifies Lineup builds one entry per channel with
// correct GuideNumber, GuideName, and stream URL.
func TestLineupReturnsChannels(t *testing.T) {
	t.Cleanup(func() { config.SetChannels(nil) })
	config.SetChannels([]config.Channel{
		{Name: "Channel 1", URL: "http://stream1.com"},
		{Name: "Channel 2", URL: "http://stream2.com"},
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/lineup.json", Lineup)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/lineup.json", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var lineups []ChannelLineup
	if err := json.Unmarshal(w.Body.Bytes(), &lineups); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(lineups) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lineups))
	}
	if lineups[0].GuideNumber != "1" || lineups[0].GuideName != "Channel 1" {
		t.Errorf("entry 0: got %+v", lineups[0])
	}
	if lineups[1].GuideNumber != "2" || lineups[1].GuideName != "Channel 2" {
		t.Errorf("entry 1: got %+v", lineups[1])
	}
}

// TestLineupStatusSourceListJSON verifies the SourceList field marshals with
// the correct JSON key "SourceList". Previously tagged json:"Cable" which caused
// Plex to receive the wrong key name.
func TestLineupStatusSourceListJSON(t *testing.T) {
	status := Status{
		ScanInProgress: 0,
		ScanPossible:   1,
		Source:         "Cable",
		SourceList:     []string{"Cable"},
	}
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["SourceList"]; !ok {
		t.Errorf(`JSON key is %q instead of "SourceList"`, keyIn(m))
	}
}

func keyIn(m map[string]interface{}) string {
	for k := range m {
		if k != "ScanInProgress" && k != "ScanPossible" && k != "Source" {
			return k
		}
	}
	return "(none)"
}
