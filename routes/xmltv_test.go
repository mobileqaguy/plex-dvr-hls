package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duncanleo/plex-dvr-hls/config"
	"github.com/gin-gonic/gin"
)

// TestXMLTVChannelIteration verifies XMLTV includes all channels from
// config.GetChannels() in its output.
func TestXMLTVChannelIteration(t *testing.T) {
	t.Cleanup(func() { config.SetChannels(nil) })
	config.SetChannels([]config.Channel{
		{Name: "News Channel", URL: "http://stream1.com"},
		{Name: "Sports Channel", URL: "http://stream2.com"},
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/xmltv", nil)

	XMLTV(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	for _, name := range []string{"News Channel", "Sports Channel"} {
		if !strings.Contains(body, name) {
			t.Errorf("expected %q in XMLTV output", name)
		}
	}
}
