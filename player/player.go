package player

import (
	"errors"
	"fmt"
	"runtime"

	"gitlab.com/GeertJohan/audio-addict/api"

	"github.com/nzlov/go-vlc"
)

var (
	ErrNoPlayer = errors.New("no player")
)

// Player manages the streaming of an AudioAddict music channel
type Player struct {
	account *api.Account

	chGetVolume  chan chan int
	chSetVolume  chan int
	chGetChannel chan chan *api.Channel
	chSetChannel chan *api.Channel

	chClose  chan struct{}
	chLock   chan struct{}
	chUnlock chan struct{}

	volume     int
	curChannel *api.Channel

	vlcInstance *vlc.Instance
	vlcPlayer   *vlc.Player

	playerStoppedHandler      func()
	playerPlayingHandler      func()
	playerTitleChangedHandler func()
	errorHandler              func(error)
}

// NewPlayer creates a new Player instance.
func NewPlayer(account *api.Account) *Player {
	p := &Player{
		account: account,

		chGetVolume:  make(chan chan int),
		chSetVolume:  make(chan int),
		chGetChannel: make(chan chan *api.Channel),
		chSetChannel: make(chan *api.Channel),

		volume: 50,

		chClose:  make(chan struct{}),
		chLock:   make(chan struct{}),
		chUnlock: make(chan struct{}),
	}

	go p.run()

	return p
}

// run manages the lifecycle of a Player in a thread-safe way.
func (p *Player) run() {
	var err error

	// Load the VLC engine with quit option
	p.vlcInstance, err = vlc.New([]string{"-q"})
	if err != nil {
		p.handleError(fmt.Errorf("New(): %v", err))
		return
	}
	defer p.vlcInstance.Release()

controlloop:
	for {
		select {
		case retCh := <-p.chGetVolume:
			if p.vlcPlayer == nil {
				retCh <- -1
				break
			}
			volume, err := p.vlcPlayer.Volume()
			if err != nil {
				p.handleError(fmt.Errorf("Volume(): %v", err))
				retCh <- -1
				break
			}
			retCh <- volume

		case volume := <-p.chSetVolume:
			p.volume = volume
			if p.vlcPlayer == nil {
				break
			}
			err := p.vlcPlayer.SetVolume(volume)
			if err != nil {
				p.handleError(fmt.Errorf("SetVolume(): %v", err))
			}

		case retCh := <-p.chGetChannel:
			retCh <- p.curChannel

		case ch := <-p.chSetChannel:
			// set current channel
			p.curChannel = ch

			// cleanup previous player (if any)
			if p.vlcPlayer != nil {
				p.vlcPlayer.Stop()
				p.vlcPlayer.Release()
			}

			streamURLs, err := ch.StreamURLs(p.account)
			if err != nil {
				p.handleError(fmt.Errorf("ch.StreamURLs(): %v", err))
				break
			}

			// Create a new media item from an url.
			media, err := p.vlcInstance.OpenMediaUri(streamURLs[0])
			if err != nil {
				p.handleError(fmt.Errorf("OpenMediaUri(): %v", err))
				break
			}

			// Create a player for the created media.
			p.vlcPlayer, err = media.NewPlayer()
			if err != nil {
				p.handleError(fmt.Errorf("NewPlayer(): %v", err))
				media.Release()
				break
			}

			// We don't need the media anymore, now that we have the player.
			media.Release()

			// set saved volume
			err = p.vlcPlayer.SetVolume(p.volume)
			if err != nil {
				p.handleError(fmt.Errorf("SetVolume: %v", err))
			}

			// get an event manager for our player.
			evt, err := p.vlcPlayer.Events()
			if err != nil {
				p.handleError(fmt.Errorf("Events(): %v", err))
				break
			}

			// Be notified when the player stops playing.
			// This is just to demonstrate usage of event callbacks.
			evt.Attach(vlc.MediaPlayerStopped, hookPlayerStoppedHandler, p)
			evt.Attach(vlc.MediaPlayerPlaying, hookPlayerPlayingHandler, p)
			evt.Attach(vlc.MediaPlayerTitleChanged, hookPlayerTitleChangedHandler, p)

			// Play the audio.
			p.vlcPlayer.Play()
			runtime.GC()
		case <-p.chClose:
			if p.vlcPlayer != nil {
				p.vlcPlayer.Stop()
				p.vlcPlayer.Release()
			}
			break controlloop
		case <-p.chLock:
			<-p.chUnlock
		}
	}

}

// used for internal locking (obtaining ownership of the player)
func (p *Player) lock() {
	p.chLock <- struct{}{}
}
func (p *Player) unlock() {
	p.chUnlock <- struct{}{}
}

// Close stops the player and closes it.
// Player instance can never be used again after calling Close().
func (p *Player) Close() {
	p.chClose <- struct{}{}
}

func (p *Player) SetPlayerStoppedHandler(handler func()) {
	p.playerStoppedHandler = handler
}

var hookPlayerStoppedHandler = func(evt *vlc.Event, data interface{}) {
	p, ok := data.(*Player)
	if !ok {
		panic("expected data to be *Player")
	}
	if p.playerStoppedHandler != nil {
		p.playerStoppedHandler()
	}
}

func (p *Player) SetPlayerPlayingHandler(handler func()) {
	p.playerPlayingHandler = handler
}

var hookPlayerPlayingHandler = func(evt *vlc.Event, data interface{}) {
	p, ok := data.(*Player)
	if !ok {
		panic("expected data to be *Player")
	}
	if p.playerPlayingHandler != nil {
		p.playerPlayingHandler()
	}
}

func (p *Player) SetPlayerTitleChangedHandler(handler func()) {
	p.playerTitleChangedHandler = handler
}

var hookPlayerTitleChangedHandler = func(evt *vlc.Event, data interface{}) {
	p, ok := data.(*Player)
	if !ok {
		panic("expected data to be *Player")
	}
	if p.playerTitleChangedHandler != nil {
		p.playerTitleChangedHandler()
	}
}

func (p *Player) SetErrorHandler(handler func(error)) {
	p.errorHandler = handler
}
func (p *Player) handleError(err error) {
	if p.errorHandler != nil {
		p.errorHandler(err)
	}
}

// PlayStop starts the player when it was stopped, and stops the player when it was started.
// When there is no player attached, false is returned.
// Otherwise the boolean indicates if the player playing after this call.
func (p *Player) PlayStop() bool {
	p.lock()
	defer p.unlock()
	//++ TODO: move this logic to run() and use channels for communication
	// 			like the other methods (SetVolume, etc)

	if p.vlcPlayer == nil {
		return false
	}

	if p.vlcPlayer.IsPlaying() {
		p.vlcPlayer.Stop()
		return false
	}
	p.vlcPlayer.Play()
	return true
}

// Play starts the player
func (p *Player) Play() bool {
	p.lock()
	defer p.unlock()
	//++ TODO: move this logic to run() and use channels for communication
	// 			like the other methods (SetVolume, etc)

	if p.vlcPlayer == nil {
		return false
	}

	p.vlcPlayer.Play()
	return true
}

// Stop stops the player. There is no pause since we're working with streams.
func (p *Player) Stop() {
	p.lock()
	defer p.unlock()
	//++ TODO: move this logic to run() and use channels for communication
	// 			like the other methods (SetVolume, etc)

	if p.vlcPlayer == nil {
		return
	}

	p.vlcPlayer.Stop()
}

// SetChannel sets the channel on the player
func (p *Player) SetChannel(c *api.Channel) {
	p.chSetChannel <- c
}

// Channel returns the current channel being played
func (p *Player) Channel() *api.Channel {
	retCh := make(chan *api.Channel)
	p.chGetChannel <- retCh
	c := <-retCh
	return c
}

// SetVolume changes the volume in the player.
// v should be a value between 0 and 100 (inclusive)
func (p *Player) SetVolume(v int) {
	p.chSetVolume <- v
}

// Volume returns the current player volume between 0 and 100.
func (p *Player) Volume() int {
	retCh := make(chan int)
	p.chGetVolume <- retCh
	volume := <-retCh
	return volume
}
