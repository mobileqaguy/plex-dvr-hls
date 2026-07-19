package routes

import (
	"bytes"
	_ "embed"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/duncanleo/plex-dvr-hls/config"
	"github.com/gin-gonic/gin"
)

//go:embed templates/xmltv.tmpl
var xmltvTemplate string

type ChannelSimplified struct {
	ID   int
	Name string
	Icon *string
}

type Programme struct {
	HourStr       string
	DateTimeStart string
	DateTimeEnd   string
}

func XMLTV(c *gin.Context) {
	var channels []ChannelSimplified

	for index, channel := range config.GetChannels() {
		channels = append(
			channels,
			ChannelSimplified{
				ID:   index + 1,
				Name: channel.Name,
				Icon: channel.Icon,
			},
		)
	}

	var programmes []Programme
	var now = time.Now()

	for i := 0; i < 24; i++ {
		var start = time.Date(now.Year(), now.Month(), now.Day(), i+1, 0, 0, 0, now.Location())
		var end = time.Date(now.Year(), now.Month(), now.Day(), i+1, 59, 59, 999, now.Location())
		var dateTimeStart = start.Format("20060102150405 -0700")

		var dateTimeEnd = end.Format("20060102150405 -0700")

		var hourStr = start.Format("3PM")

		programmes = append(
			programmes,
			Programme{
				HourStr:       hourStr,
				DateTimeStart: dateTimeStart,
				DateTimeEnd:   dateTimeEnd,
			},
		)
	}

	t, err := template.New("xmltv.tmpl").Parse(xmltvTemplate)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var b bytes.Buffer
	err = t.Execute(
		&b,
		gin.H{
			"channels":   channels,
			"programmes": programmes,
		},
	)

	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/xml", b.Bytes())
}
