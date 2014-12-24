// // Direct port of http://wiki.videolan.org/LibVLC_Tutorial
// //
// // Run this program with -v, -vv or -vvv parameters for progressively more
// // verbose debug output from libvlc.
package main

// import (
// 	"fmt"
// 	vlc "github.com/jteeuwen/go-vlc"
// 	"os"
// )

// const uri = "http://prem2.di.fm:80/electrohouse?2519cde93c2cf51"

// func main() {
// 	var inst *vlc.Instance
// 	var player *vlc.Player
// 	var media *vlc.Media
// 	var evt *vlc.EventManager
// 	var err error

// 	// Load the VLC engine.
// 	if inst, err = vlc.New(os.Args); err != nil {
// 		fmt.Fprintf(os.Stderr, "[e] New(): %v", err)
// 		return
// 	}

// 	defer inst.Release()

// 	// Create a new media item from an url.
// 	if media, err = inst.OpenMediaUri(uri); err != nil {
// 		fmt.Fprintf(os.Stderr, "[e] OpenMediaUri(): %v", err)
// 		return
// 	}

// 	// Create a player for the created media.
// 	if player, err = media.NewPlayer(); err != nil {
// 		fmt.Fprintf(os.Stderr, "[e] NewPlayer(): %v", err)
// 		media.Release()
// 		return
// 	}

// 	defer player.Release()

// 	// We don't need the media anymore, now that we have the player.
// 	media.Release()
// 	media = nil

// 	// get an event manager for our player.
// 	if evt, err = player.Events(); err != nil {
// 		fmt.Fprintf(os.Stderr, "[e] Events(): %v", err)
// 		return
// 	}

// 	// Be notified when the player stops playing.
// 	// This is just to demonstrate usage of event callbacks.
// 	evt.Attach(vlc.MediaPlayerStopped, handler, "wahey!")

// 	// Play the audio.
// 	player.Play()

// 	// for {
// 	// 	line, err := linenoise.Line("> ")
// 	// 	if err != nil {
// 	// 		log.Fatalf("error reading line: %v", err)
// 	// 	}
// 	// 	switch line {
// 	// 	case "stop":
// 	// 		// Stop playing.
// 	// 		player.Stop()
// 	// 	case "quit":
// 	// 		os.Exit(1)
// 	// 	}
// 	// }

// 	select {}
// }

// func handler(evt *vlc.Event, data interface{}) {
// 	fmt.Printf("[i] %s occurred: %s\n", evt.Type, data.(string))
// }
