package main

/*
#cgo LDFLAGS: -l g15daemon_client -l g15render -Wl,--allow-multiple-definition
#include <g15daemon_client.h>
#include <libg15.h>
#include <libg15render.h>
#include <stdlib.h>
#include "repeat.c"
*/
import "C"

import (
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/godbus/dbus/v5"
)

// MusicMetadata contains all the data given from a spotify dbus message
type MusicMetadata struct {
	TrackID          string
	Length           int64
	ArtURL           string
	Album            string
	AlbumArtist      []string
	Artist           []string
	AutoRating       float64
	DiscNumber       int32
	Title            string
	TrackNumber      uint32
	URL              string
	PlaybackStatus   string // Can be Paused, Playing, Stopped
	PlaybackPosition int64
        PlaybackShuffle bool
        PlaybackLoop string
}

var (
	canvas       *C.struct_g15canvas
	g15screen_fd C.int
)

func getMetadata(s dbus.BusObject) (MusicMetadata) {
	metadatacontainer, err := s.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		log.Fatal(err)
	}
	playbackposition, err := s.GetProperty("org.mpris.MediaPlayer2.Player.Position")
	log.Printf("playbackposition: %v", playbackposition.Value())
	if err != nil {
		log.Fatal(err)
	}
	playbackloop, err := s.GetProperty("org.mpris.MediaPlayer2.Player.LoopStatus")
	if err != nil {
		log.Fatal(err)
	}
	playbackstatus, err := s.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		log.Fatal(err)
	}
	metadata, _ := metadatacontainer.Value().(map[string]dbus.Variant)
	//  length := metadata["mpris:length"].value

	if playbackstatus.Value().(string) == "Stopped" {
		return MusicMetadata{
		PlaybackStatus:   playbackstatus.Value().(string),

                }

	}

	return MusicMetadata{
		TrackID:          metadata["mpris:trackid"].Value().(string),
		Length:           metadata["mpris:length"].Value().(int64),
		ArtURL:           metadata["mpris:artUrl"].Value().(string),
		Album:            metadata["xesam:album"].Value().(string),
		AlbumArtist:      metadata["xesam:albumArtist"].Value().([]string),
		Artist:           metadata["xesam:artist"].Value().([]string),
		AutoRating:       metadata["xesam:autoRating"].Value().(float64),
		DiscNumber:       metadata["xesam:discNumber"].Value().(int32),
		Title:            metadata["xesam:title"].Value().(string),
		TrackNumber:      metadata["xesam:trackNumber"].Value().(uint32),
		URL:              metadata["xesam:url"].Value().(string),
		PlaybackStatus:   playbackstatus.Value().(string),
		PlaybackPosition: playbackposition.Value().(int64),
                PlaybackLoop: playbackloop.Value().(string),

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

func drawCentered(c *C.struct_g15canvas, text string, row int, size int) {
	textpixelwidth := 0
	switch size {
	case C.G15_TEXT_MED:
		textpixelwidth = 5
	case C.G15_TEXT_LARGE:
		textpixelwidth = 8
	case C.G15_TEXT_SMALL:
		textpixelwidth = 4
	}
	pos := (C.G15_LCD_WIDTH - (textpixelwidth * len(text))) / 2
	//fmt.Printf("pos: %d, len(text): %d, text: %s\n", pos, len(text), text)
	if pos < 0 {
		pos = 0
	}
	C.g15r_renderString(c, (*C.uchar)((unsafe.Pointer)(C.CString(text))), C.int(row), C.int(size), C.uint(uint(pos)), 0)

}

func screenDraw(md MusicMetadata) {
	C.g15r_clearScreen(canvas, 0)
	if md.PlaybackStatus == "Stopped" {
		drawCentered(canvas, "Music Playback Stopped", 2, C.G15_TEXT_MED)
	} else {
		drawCentered(canvas, md.Title, 0, C.G15_TEXT_MED)
		drawCentered(canvas, md.Album, 1, C.G15_TEXT_MED)
		drawCentered(canvas, md.Artist[0], 2, C.G15_TEXT_MED)
                //log.Printf("length: %d, pos: %d, a: %d\n", md.Length, md.PlaybackPosition, int64(md.Length)-md.PlaybackPosition)
                timestring := time.Duration((md.Length-md.PlaybackPosition*1000)*1000).Truncate(time.Second).String()
                //log.Printf("timestring: %s", timestring)
		drawCentered(canvas, timestring, 4, C.G15_TEXT_MED)

                switch md.PlaybackLoop {
                case "None":
                  C.g15r_pixelOverlay(canvas,C.G15_LCD_WIDTH-C.REPEAT_FRAME_WIDTH-8, C.G15_LCD_HEIGHT-C.REPEAT_FRAME_HEIGHT,C.REPEAT_FRAME_WIDTH,C.REPEAT_FRAME_HEIGHT, &C.repeat_frame_data[2][0])
                case "Track":
                  C.g15r_pixelOverlay(canvas,C.G15_LCD_WIDTH-C.REPEAT_FRAME_WIDTH-8, C.G15_LCD_HEIGHT-C.REPEAT_FRAME_HEIGHT,C.REPEAT_FRAME_WIDTH,C.REPEAT_FRAME_HEIGHT, &C.repeat_frame_data[1][0])
                case "Playlist":
                  C.g15r_pixelOverlay(canvas,C.G15_LCD_WIDTH-C.REPEAT_FRAME_WIDTH-8, C.G15_LCD_HEIGHT-C.REPEAT_FRAME_HEIGHT,C.REPEAT_FRAME_WIDTH,C.REPEAT_FRAME_HEIGHT, &C.repeat_frame_data[0][0])

                }

	}

	for C.g15_send(g15screen_fd, (*C.char)(unsafe.Pointer(&canvas.buffer)), C.G15_BUFFER_LEN) < 0 {
		log.Fatal("RIP")
	}

}

func main() {
	dconn, err := dbus.SessionBus()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Initializing screen")
	screenInit()
	log.Printf("Initialized screen")
	player := os.Args[1]
	log.Printf("connecting to dbus org.mpris.MediaPlayer2.%s", player)
	obj := dconn.Object("org.mpris.MediaPlayer2."+player, "/org/mpris/MediaPlayer2")
	for {
		screenDraw(getMetadata(obj))
		time.Sleep(time.Second)
	}
}
