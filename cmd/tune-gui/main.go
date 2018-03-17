package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/GeertJohan/audio-addict/aaplayer"
	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/mitchellh/panicwrap"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/quickcontrols2"

	"github.com/GeertJohan/tune/api"
	"github.com/GeertJohan/tune/config"
)

//go:generate qtmoc
type ChannelBridge struct {
	core.QObject

	_ func()                                                                              `signal:"clearChannels"`
	_ func(channelKey string, channelName string, channelImage string, trackTitle string) `signal:"addChannel"`
}

type PlayerBridge struct {
	core.QObject
}

func main() {
	// Add panicwrap to catch any panics and save them to file.
	exitStatus, err := panicwrap.BasicWrap(panicToFile)
	if err != nil {
		panic(err)
	}
	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}

	// Enable high dpi scaling
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)

	// Init the application in Qt
	// TODO: move args to seperate --qt-args param so we can have our own args without disrupting qt and vice versa.
	gui.NewQGuiApplication(len(os.Args), os.Args)

	// Use the material style for qml controls, we just have to set some colors.
	quickcontrols2.QQuickStyle_SetStyle("material")

	// Create the QML application engine
	app := qml.NewQQmlApplicationEngine(nil)

	// create a new ChannelBridge for interfacing back an forth channel information
	channelBridge := NewChannelBridge(nil)
	app.RootContext().SetContextProperty("channelBridge", channelBridge)
	playerBridge := NewPlayerBridge(nil)
	app.RootContext().SetContextProperty("playerBridge", playerBridge)

	// load the main qml file
	app.Load(core.NewQUrl3("qrc:///qml/aagui.qml", 0))

	var chGuiClosed = make(chan chan struct{})

	// hardcode di.fm network for now
	network := api.NetworkDI
	conf := loadConfig()
	// load account
	account := loadAccount(network, conf)
	streamList := loadStreamList(network, conf, account)

	go startChannelList(streamList, channelBridge)
	go startPlayer(chGuiClosed, conf, account, playerBridge)

	// enter the main event loop, blocks until gui is exited
	gui.QGuiApplication_Exec()
	// indicate gui closed
	var chPlayerClosed = make(chan struct{})
	chGuiClosed <- chPlayerClosed
	// wait until player closed
	<-chPlayerClosed
}

func loadConfig() *config.Config {
	// Load aa configuration. Note this config is compatible with aacli and by default is saved in the same location.
	conf, err := config.Load()
	if err != nil {
		if err == config.ErrNoConfigFile {
			conf = config.New()
		} else {
			fmt.Printf("error loading conf: %v\n", err)
			os.Exit(1)
		}
	}
	return conf
}

func loadAccount(network *api.Network, conf *config.Config) *api.Account {
	// authenticate account
	var err error
	var account *api.Account
	if conf.Account.APIKey != "" {
		account, err = network.AuthenticateAPIKey(conf.Account.APIKey)
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
			conf.Account.APIKey = account.APIKey
			conf.Save()
			break
		}
	}
	return account
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

func loadStreamList(network *api.Network, conf *config.Config, account *api.Account) *api.Streamlist {
	var err error
	var streamList *api.Streamlist
	if conf.Settings.StreamlistKey != "" {
		streamList, err = network.StreamlistByKey(conf.Settings.StreamlistKey)
		if err != nil {
			fmt.Println("Could not use saved stream quality, selecting best quality.")
			linenoise.Line("Press enter to continue")
		}
	}
	if streamList == nil {
		streamList = network.BestStreamlist(account.Premium)
		conf.Settings.StreamlistKey = streamList.Key
		conf.Save()
	}
	return streamList
}

func startChannelList(streamList *api.Streamlist, channelBridge *ChannelBridge) {
	// get all channels
	channels, err := streamList.Channels()
	if err != nil {
		fmt.Printf("error getting channels: %v\n", err)
		os.Exit(1)
	}
	channelsByKey := make(map[string]*api.Channel)
	for _, channel := range channels {
		channelsByKey[channel.Key] = channel
	}

	trackHistory, err := streamList.Network.TrackHistory()
	if err != nil {
		fmt.Printf("error getting track history: %v\n", err)
		os.Exit(1)
	}
	// Clean the exiting channel list in gui
	channelBridge.ClearChannels()
	// Add channels to GUI via bridge
	for _, ch := range channels {
		trackInfo := trackHistory[strconv.Itoa(ch.ID)]
		channelBridge.AddChannel(ch.Key, ch.Name, trackInfo.ArtURLHTTPS(), trackInfo.Name)
	}
}

func startPlayer(chGuiClosed chan chan struct{}, conf *config.Config, account *api.Account, playerBridge *PlayerBridge) {

	// // create a clock
	// var clk *clock.Clock
	// {
	// 	ping, err := api.Ping()
	// 	if err != nil {
	// 		fmt.Printf("error pinging audioaddict server: %v\n", err)
	// 		os.Exit(1)
	// 	}
	// 	clk = clock.New(ping.Time)
	// }

	// create an aaplayer
	player := aaplayer.NewPlayer(account)
	// restore default config
	player.SetVolume(conf.Player.Volume)
	// defer proper shutdown of the player
	defer player.Close()

	// 	var updateTrack func()
	// 	updateTrack = func() {
	// 		track, err := player.Channel().CurrentTrack()
	// 		if err != nil {
	// 			display.Notify(fmt.Sprintf("error getting title: %v", err))
	// 		} else {
	// 			display.SetTrackTitle(track.Name)
	// 		}
	// 		duration := time.Duration(track.Duration) * time.Second
	// 		started := time.Unix(int64(track.Started), 0)
	// 		now := clk.Now()
	// 		passed := now.Sub(started)
	// 		display.SetTrackDuration(duration, passed)
	// 		go func() {
	// 			left := duration - passed
	// 			if left < 2*time.Second {
	// 				// avoid infinite loop when information provided by api is incorrect
	// 				left = 2 * time.Second
	// 			}
	// 			<-time.After(left)
	// 			updateTrack()
	// 		}()
	// 	}
	// 	player.SetPlayerStoppedHandler(func() {
	// 		display.Notify("playback stopped")
	// 		display.SetPlaying(false)
	// 		display.SetTrackTitle("N/A")
	// 	})
	// 	player.SetPlayerPlayingHandler(func() {
	// 		display.Notify("playback started")
	// 		display.SetPlaying(true)
	//
	// 		updateTrack()
	// 	})
	// 	player.SetPlayerTitleChangedHandler(updateTrack)
	// 	player.SetErrorHandler(func(err error) {
	// 		display.Notify(fmt.Sprintf("error: %v", err))
	// 	})
	//
	// 	// convert blocking call termbox.PollEvent() to channel send
	// 	eventChan := make(chan termbox.Event)
	// 	go func() {
	// 		for {
	// 			event := termbox.PollEvent()
	// 			eventChan <- event
	// 		}
	// 	}()
	//
	// 	// catch interrupt and kill signals from the os
	// 	sigChan := make(chan os.Signal)
	// 	signal.Notify(sigChan, os.Interrupt)
	// 	signal.Notify(sigChan, os.Kill)
	//
	// 	// start channel that was previously being played
	// 	if conf.Player.LastPlayedChannel != "" {
	// 		channel := channelsByKey[conf.Player.LastPlayedChannel]
	// 		if channel != nil {
	// 			player.SetChannel(channel)
	// 			display.SetChannel(channel.Name, channel.Key)
	// 		}
	// 	}
	//
	// 	changeVolume := func(change int) {
	// 		conf.Player.Volume += change
	// 		if conf.Player.Volume < 0 {
	// 			conf.Player.Volume = 0
	// 		}
	// 		if conf.Player.Volume > 100 {
	// 			conf.Player.Volume = 100
	// 		}
	// 		// first set volume (fast audio feedback to user), afterwards save conf to disk
	// 		player.SetVolume(conf.Player.Volume)
	// 		display.Notify(fmt.Sprintf("volume set to %02d%%", conf.Player.Volume))
	// 		conf.Save()
	// 	}
	//
	// eventloop:
	// 	for {
	// 		select {
	// 		case event := <-eventChan:
	// 			// switch on event type
	// 			switch event.Type {
	// 			case termbox.EventKey: // actions depend on key
	// 				switch event.Key {
	// 				case termbox.KeyCtrlZ, termbox.KeyCtrlC, termbox.KeyEsc:
	// 					break eventloop
	// 				case termbox.KeySpace:
	// 					player.PlayStop()
	// 				case termbox.KeyArrowUp:
	// 					display.MoveChannelListSelection(-1)
	// 				case termbox.KeyArrowDown:
	// 					display.MoveChannelListSelection(1)
	// 				case termbox.KeyEnter:
	// 					channelKey := display.GetChannelSelection()
	// 					ch := channelsByKey[channelKey]
	// 					player.SetChannel(ch)
	// 					display.SetChannel(ch.Name, ch.Key)
	// 					conf.Player.LastPlayedChannel = ch.Key
	// 					conf.Save()
	// 				}
	// 				switch event.Ch {
	// 				case 'q':
	// 					break eventloop
	// 				case 'p':
	// 					player.PlayStop()
	// 				case '-', '_':
	// 					changeVolume(-5)
	// 				case '+', '=':
	// 					changeVolume(5)
	// 				}
	//
	// 			case termbox.EventResize:
	// 				display.Resize(event.Width, event.Height)
	//
	// 			case termbox.EventError: // quit
	// 				// fmt.Printf("quitting because of termbox error: %v", event.Err)
	// 				break eventloop
	// 			}
	// 		case _ = <-sigChan:
	// 			break eventloop
	// 		}
	// 	}

	// wait until GUI is closed
	chPlayerClosed := <-chGuiClosed
	// TODO: do player shutdown stuff
	// indicate player has closed
	close(chPlayerClosed)
}
