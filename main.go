package main

/*
#cgo LDFLAGS: -l g15daemon_client -l g15render
#include <g15daemon_client.h>
#include <libg15.h>
#include <libg15render.h>
#include <stdlib.h>
*/
import "C"

import (
	"github.com/godbus/dbus"
	"log"
	"fmt"
	"unsafe"
)

// MusicMetadata contains all the data given from a spotify dbus message
type MusicMetadata struct {
	TrackID          string
	Length           uint64
	ArtURL           string
	Album            string
	AlbumArtist      []string
	Artist           []string
	AutoRating       float64
	DiscNumber       int32
	Title            string
	TrackNumber      int32
	URL              string
	PlaybackStatus   string // Can be Paused, Playing, Stopped
	PlaybackPosition uint64
}

var (
	canvas *C.struct_g15canvas
	g15screen_fd C.int
)

func getMetadata(s *dbus.Signal) MusicMetadata {
	body := s.Body
	metadatacontainer := body[1].(map[string]dbus.Variant)
	metadata := metadatacontainer["Metadata"].Value().(map[string]dbus.Variant)
	playbackStatus := metadatacontainer["PlaybackStatus"].Value()
	//  length := metadata["mpris:length"].value

	return MusicMetadata{
		TrackID:        metadata["mpris:trackid"].Value().(string),
		Length:         metadata["mpris:length"].Value().(uint64),
		ArtURL:         metadata["mpris:artUrl"].Value().(string),
		Album:          metadata["xesam:album"].Value().(string),
		AlbumArtist:    metadata["xesam:albumArtist"].Value().([]string),
		Artist:         metadata["xesam:artist"].Value().([]string),
		AutoRating:     metadata["xesam:autoRating"].Value().(float64),
		DiscNumber:     metadata["xesam:discNumber"].Value().(int32),
		Title:          metadata["xesam:title"].Value().(string),
		TrackNumber:    metadata["xesam:trackNumber"].Value().(int32),
		URL:            metadata["xesam:url"].Value().(string),
		PlaybackStatus: playbackStatus.(string),
	}

}

func screenInit() {
	g15screen_fd = C.new_g15_screen(C.G15_G15RBUF)
	if g15screen_fd < 0 {
		log.Fatal("Can't connect to the G15daemon")
	}
	canvas = (*C.struct_g15canvas)(C.malloc(C.sizeof_struct_g15canvas))
	C.g15r_initCanvas(canvas)
}

func screenDraw(md MusicMetadata) {
  C.g15r_renderString(canvas, (*C.uchar)((unsafe.Pointer)(C.CString(fmt.Sprintf("Title: %s", md.Title)))),0, C.G15_TEXT_MED,0,0)
  C.g15r_renderString(canvas, (*C.uchar)((unsafe.Pointer)(C.CString(fmt.Sprintf("Album: %s", md.Album)))),1, C.G15_TEXT_MED,0,0)
  C.g15r_renderString(canvas, (*C.uchar)((unsafe.Pointer)(C.CString(fmt.Sprintf("Artist: %s", md.Artist[0])))),2, C.G15_TEXT_MED,0,0)
  C.g15r_renderString(canvas, (*C.uchar)((unsafe.Pointer)(C.CString(fmt.Sprintf("Track: %d", md.TrackNumber)))),3, C.G15_TEXT_MED,0,0)

  for C.g15_send(g15screen_fd, (*C.char)(unsafe.Pointer(&canvas.buffer)),C.G15_BUFFER_LEN)<0 {
		log.Fatal("RIP")
  }

}

func main() {
	dconn, err := dbus.SessionBus()
	if err != nil {
		log.Fatal(err)
	}
	dconn.AddMatchSignal(dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))
	sigchan := make(chan *dbus.Signal)
	dconn.Signal(sigchan)
	log.Print("Registered Signal Handler")
	screenInit()
	for {
		rs := <-sigchan
		screenDraw(getMetadata(rs))

	}
}
