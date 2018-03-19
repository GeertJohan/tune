package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/GeertJohan/go.linenoise"
	"github.com/mitchellh/panicwrap"
	"github.com/nsf/termbox-go"

	"github.com/GeertJohan/tune/api"
	"github.com/GeertJohan/tune/clock"
	tuneplayer "github.com/GeertJohan/tune/player"
	tunesettings "github.com/GeertJohan/tune/settings"
)

func main() {
	exitStatus, err := panicwrap.BasicWrap(panicToFile)
	if err != nil {
		panic(err)
	}
	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}

	settings, err := tunesettings.Load()
	if err != nil {
		fmt.Printf("error loading or creating settings file: %v\n", err)
		os.Exit(1)
	}

	// hardcode di.fm network for now
	network := api.NetworkDI

	// authenticate account
	var account *api.Account
	if settings.Account.APIKey != "" {
		account, err = network.AuthenticateAPIKey(settings.Account.APIKey)
		if err == api.ErrInvalidCredentials {
			fmt.Println("Could not authenticate with saved API key.")
		} else if err != nil {
			fmt.Printf("error authenticating with API key: %v\n", err)
			os.Exit(1)
		}
	}
	if account == nil {
		fmt.Println("Please enter your AudioAddict username and password.")
		for {
			username := mustReadLine("username: ")
			password := mustReadLine("password: ")
			account, err = network.AuthenticateUserPass(username, password)
			if err == api.ErrInvalidCredentials {
				fmt.Println("Invalid username and/or password, please try again.")
				continue
			} else if err != nil {
				fmt.Printf("error authenticating with username/password: %v\n", err)
				os.Exit(1)
			}
			settings.Account.APIKey = account.APIKey
			settings.Save()
			break
		}
	}

	var sl *api.Streamlist
	if settings.Settings.StreamlistKey != "" {
		sl, err = network.StreamlistByKey(settings.Settings.StreamlistKey)
		if err != nil {
			fmt.Println("Could not use saved stream quality, selecting best quality.")
			linenoise.Line("Press enter to continue")
		}
	}
	if sl == nil {
		sl = network.BestStreamlist(account.Premium)
		settings.Settings.StreamlistKey = sl.Key
		settings.Save()
	}

	// get all channels
	channels, err := sl.Channels()
	if err != nil {
		fmt.Printf("error getting channels: %v\n", err)
		os.Exit(1)
	}
	channelsByKey := make(map[string]*api.Channel)
	for _, channel := range channels {
		channelsByKey[channel.Key] = channel
	}

	// create a clock
	var clk *clock.Clock
	{
		ping, err := api.Ping()
		if err != nil {
			fmt.Printf("error pinging audioaddict server: %v\n", err)
			os.Exit(1)
		}
		clk = clock.New(ping.Time)
	}

	// create display
	display, err := NewDisplay(network.Name)
	if err != nil {
		fmt.Printf("error setting up display: %v\n", err)
		os.Exit(1)
	}
	defer display.Close()

	// setup tracklist on display
	go func() {
		trackHistory, err := network.TrackHistory()
		if err != nil {
			fmt.Printf("error getting track history: %v\n", err)
			os.Exit(1)
		}
		channelList := make([]*channelInfo, 0, len(channels))
		for _, ch := range channels {
			ci := &channelInfo{
				channelKey:  ch.Key,
				channelName: ch.Name,
			}

			trackInfo := trackHistory[strconv.Itoa(ch.ID)]
			if trackInfo != nil {
				ci.trackTitle = trackInfo.Name
			}
			channelList = append(channelList, ci)
		}
		display.SetChannelList(channelList)
	}()

	// create player
	player := tuneplayer.NewPlayer(account)
	player.SetVolume(settings.Player.Volume)
	defer player.Close()

	var updateTrack func()
	updateTrack = func() {
		track, err := player.Channel().CurrentTrack()
		if err != nil {
			display.Notify(fmt.Sprintf("error getting title: %v", err))
		} else {
			display.SetTrackTitle(track.Name)
		}
		duration := time.Duration(track.Duration) * time.Second
		started := time.Unix(int64(track.Started), 0)
		now := clk.Now()
		passed := now.Sub(started)
		display.SetTrackDuration(duration, passed)
		go func() {
			left := duration - passed
			if left < 2*time.Second {
				// avoid infinite loop when information provided by api is incorrect
				left = 2 * time.Second
			}
			<-time.After(left)
			updateTrack()
		}()
	}
	player.SetPlayerStoppedHandler(func() {
		display.Notify("playback stopped")
		display.SetPlaying(false)
		display.SetTrackTitle("N/A")
	})
	player.SetPlayerPlayingHandler(func() {
		display.Notify("playback started")
		display.SetPlaying(true)

		updateTrack()
	})
	player.SetPlayerTitleChangedHandler(updateTrack)
	player.SetErrorHandler(func(err error) {
		display.Notify(fmt.Sprintf("error: %v", err))
	})

	// convert blocking call termbox.PollEvent() to channel send
	eventChan := make(chan termbox.Event)
	go func() {
		for {
			event := termbox.PollEvent()
			eventChan <- event
		}
	}()

	// catch interrupt and kill signals from the os
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	// start channel that was previously being played
	if settings.Player.LastPlayedChannel != "" {
		channel := channelsByKey[settings.Player.LastPlayedChannel]
		if channel != nil {
			player.SetChannel(channel)
			display.SetChannel(channel.Name, channel.Key)
		}
	}

	changeVolume := func(change int) {
		settings.Player.Volume += change
		if settings.Player.Volume < 0 {
			settings.Player.Volume = 0
		}
		if settings.Player.Volume > 100 {
			settings.Player.Volume = 100
		}
		// first set volume (fast audio feedback to user), afterwards save conf to disk
		player.SetVolume(settings.Player.Volume)
		display.Notify(fmt.Sprintf("volume set to %02d%%", settings.Player.Volume))
		settings.Save()
	}

eventloop:
	for {
		select {
		case event := <-eventChan:
			// switch on event type
			switch event.Type {
			case termbox.EventKey: // actions depend on key
				switch event.Key {
				case termbox.KeyCtrlZ, termbox.KeyCtrlC, termbox.KeyEsc:
					break eventloop
				case termbox.KeySpace:
					channelKey := display.GetChannelSelection()
					if player.Channel() != nil && player.Channel().Key == channelKey {
						player.PlayStop()
					} else {
						ch := channelsByKey[channelKey]
						player.SetChannel(ch)
						display.SetChannel(ch.Name, ch.Key)
						settings.Player.LastPlayedChannel = ch.Key
						settings.Save()
					}
				case termbox.KeyArrowUp:
					display.MoveChannelListSelection(-1)
				case termbox.KeyArrowDown:
					display.MoveChannelListSelection(1)
				}

				switch event.Ch {
				case 'q':
					break eventloop
				case '-', '_':
					changeVolume(-5)
				case '+', '=':
					changeVolume(5)
				}

			case termbox.EventResize:
				display.Resize(event.Width, event.Height)

			case termbox.EventError: // quit
				// fmt.Printf("quitting because of termbox error: %v", event.Err)
				break eventloop
			}
		case _ = <-sigChan:
			break eventloop
		}
	}
}

func mustReadLine(prompt string) string {
	for {
		line, err := linenoise.Line(prompt)
		if err != nil {
			fmt.Printf("error reading line: %v\n", err)
			os.Exit(1)
		}
		if len(line) != 0 {
			return line
		}
	}
}
