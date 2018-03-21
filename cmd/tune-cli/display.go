package main

import (
	"fmt"
	"time"

	"github.com/foize/go.fifo"
	"github.com/nsf/termbox-go"
)

const (
	colorDefaultForeground    = termbox.Attribute(249)
	colorHelpForeground       = termbox.Attribute(245)
	colorTrackTitleBackground = termbox.Attribute(245)
	colorBlack                = termbox.ColorBlack
)

const (
	runePlaying       = '▶'
	runeStopped       = '◾'
	runeTimebarStart  = '['
	runeTimebarPassed = '#'
	runeTimebarLeft   = '-'
	runeTimebarEnd    = ']'
	runeSelected      = '→'
)

type Display struct {
	volume int

	title               string // window title
	channelKey          string // channel key
	channelName         string // channel name
	playing             bool   // indicates if player is currently playing
	trackTitle          string // track title
	trackDuration       time.Duration
	trackPassed         time.Duration
	message             string // notification message
	channelList         []*channelInfo
	channelListSelected int
	channelListStart    int

	chStop   chan struct{}
	chLock   chan struct{}
	chUnlock chan struct{}
	chNotify chan string

	size struct {
		x int
		y int
	}
}

type channelInfo struct {
	channelKey  string
	channelName string
	trackTitle  string
}

func NewDisplay() (*Display, error) {
	// init display
	err := termbox.Init()
	if err != nil {
		return nil, fmt.Errorf("termbox.Init(): %v", err)
	}
	termbox.HideCursor()
	err = termbox.Clear(colorBlack, colorBlack)
	if err != nil {
		return nil, fmt.Errorf("termbox.Clear(): %v", err)
	}
	termbox.SetOutputMode(termbox.Output256)
	d := &Display{
		chStop:   make(chan struct{}),
		chLock:   make(chan struct{}),
		chUnlock: make(chan struct{}),
		chNotify: make(chan string),
	}

	go d.run()
	go d.notificationManager()
	d.Resize(termbox.Size())

	return d, nil
}

func (d *Display) Close() {
	d.chStop <- struct{}{}
	termbox.Clear(colorBlack, colorBlack)
	termbox.Close()
}

func (d *Display) SetTitle(title string) {
	d.title = title
	d.drawBasics()
}

func (d *Display) run() {
	sec := 1 * time.Second
	ticker := time.NewTicker(sec)

	for {
		select {
		case <-d.chStop:
			return
		case <-ticker.C:
			if d.playing {
				d.trackPassed += sec
				d.drawTime()
				termbox.Flush()
			}
		case <-d.chLock:
			<-d.chUnlock
		}
	}
}
func (d *Display) lock() {
	d.chLock <- struct{}{}
}
func (d *Display) unlock() {
	d.chUnlock <- struct{}{}
}

func (d *Display) notificationManager() {
	//++ TODO: the notificationManager can easily put multiple notifications on one line...
	// also: it doesn't always make sense to display a notification for very long.. for instance volume 50, 55, 60, 65.. etc..
	msgQueue := fifo.NewQueue()

	for {
		var message string
		{
			msgItem := msgQueue.Next()
			if msgItem != nil {
				message = msgItem.(string)
				goto display
			}
		}
		message = <-d.chNotify

	display:
		d.lock()
		prevLength := len(d.message)
		d.message = message
		y := d.size.y - 2
		for i, c := range d.message {
			x := i + len(d.title) + 3
			if x > d.size.x {
				break
			}
			termbox.SetCell(x, y, c, colorDefaultForeground, colorBlack)
		}
		for i := len(d.message); i < prevLength; i++ {
			x := i + len(d.title) + 3
			termbox.SetCell(x, y, ' ', colorBlack, colorBlack)
		}
		termbox.Flush()
		d.unlock()

		// wait before showing next item, until then queue all notifications
		notificationTimeout := time.After(150 * time.Millisecond)
	waitLoop:
		for {
			select {
			case <-notificationTimeout:
				break waitLoop
			case msg := <-d.chNotify:
				msgQueue.Add(msg)
			}
		}
	}
}

func (d *Display) drawBasics() {
	// title and title seperator
	d.writeText(d.title+` - `, 0, 0, colorBlack, termbox.Attribute(249))
	d.writeText(d.title+` - `, 0, d.size.y-2, colorDefaultForeground, colorBlack)

	// display key help
	helpmessage := `q: quit  up/down: select channel  space: play/stop  +/-: volume`
	for i, c := range helpmessage {
		termbox.SetCell(i, d.size.y-1, c, colorHelpForeground, colorBlack)
	}
}
func (d *Display) drawChannel() {
	x := len(d.title) + 3
	x = d.writeText(d.channelName, x, 0, colorBlack, termbox.Attribute(249))
	d.clearRow(x, 0, termbox.Attribute(249))
}
func (d *Display) drawTrackTitle() {
	x := 5
	x = d.writeText(d.trackTitle, x, 1, colorTrackTitleBackground, colorBlack)
	d.clearRow(x, 1, colorBlack)
}
func (d *Display) drawPlaying() {
	if d.playing {
		termbox.SetCell(2, 1, runePlaying, termbox.ColorGreen, colorBlack)
	} else {
		termbox.SetCell(2, 1, runeStopped, termbox.ColorRed, colorBlack)
	}
}
func (d *Display) drawTime() {
	passedMins := int(d.trackPassed.Minutes())
	passedSecs := int(d.trackPassed.Seconds())
	passedStr := fmt.Sprintf(`%02d:%02d`, passedMins, passedSecs%60)
	durationMins := int(d.trackDuration.Minutes())
	durationSecs := int(d.trackDuration.Seconds())
	durationStr := fmt.Sprintf(`%02d:%02d`, durationMins, durationSecs%60)

	offsetStart := len(passedStr)
	for x, c := range passedStr {
		termbox.SetCell(x, 2, c, colorDefaultForeground, colorBlack)
	}

	offsetEnd := len(durationStr)
	for x, c := range durationStr {
		termbox.SetCell(d.size.x-offsetEnd+x, 2, c, colorDefaultForeground, colorBlack)
	}
	posStart := offsetStart + 1
	posEnd := d.size.x - offsetEnd - 2
	barSize := posEnd - posStart - 2
	if durationSecs < 1 {
		termbox.Flush()
		return
	}
	barSizePassed := passedSecs * barSize / durationSecs
	if barSizePassed < 1 {
		barSizePassed = 1
	}
	if barSizePassed > barSize {
		barSizePassed = barSize
	}
	termbox.SetCell(posStart, 2, runeTimebarStart, colorDefaultForeground, colorBlack)
	termbox.SetCell(posEnd, 2, runeTimebarEnd, colorDefaultForeground, colorBlack)
	for x := posStart + 1; x <= posStart+1+barSizePassed; x++ {
		termbox.SetCell(x, 2, runeTimebarPassed, colorDefaultForeground, colorBlack)
	}
	for x := posStart + 1 + barSizePassed + 1; x < posEnd; x++ {
		termbox.SetCell(x, 2, runeTimebarLeft, colorDefaultForeground, colorBlack)
	}
}

func (d *Display) drawVolume() {}

func (d *Display) drawChannelList() {
	if len(d.channelList) == 0 {
		return
	}
	viewHeight := d.size.y - 7
	viewStartY := 4

	if d.channelListStart > d.channelListSelected {
		d.channelListStart = d.channelListSelected
	} else if d.channelListStart+viewHeight <= d.channelListSelected {
		d.channelListStart = d.channelListSelected - viewHeight + 1
	}
	if d.channelListStart+viewHeight > len(d.channelList) {
		d.channelListStart = len(d.channelList) - viewHeight
	}
	if d.channelListStart < 0 {
		d.channelListStart = 0
	}

	for i := 0; i < viewHeight; i++ {
		channelListPosition := i + d.channelListStart
		y := viewStartY + i
		x := 5
		if channelListPosition >= len(d.channelList) {
			d.clearRow(x, y, colorBlack)
			continue
		}
		chinfo := d.channelList[channelListPosition]
		selectionRune := ' '
		attrBackground := colorBlack
		if d.channelListSelected == channelListPosition {
			selectionRune = runeSelected
			attrBackground = termbox.Attribute(235)
		}
		termbox.SetCell(3, y, selectionRune, termbox.ColorYellow, colorBlack)
		if d.playing && d.channelKey == chinfo.channelKey {
			x = d.writeText(` `+string(runePlaying)+` `, x, y, termbox.ColorGreen, colorBlack)
		}
		x = d.writeText(chinfo.channelName+`: `, x, y, colorDefaultForeground|termbox.AttrBold, attrBackground)
		x = d.writeText(chinfo.trackTitle, x, y, colorDefaultForeground, attrBackground)
		d.clearRow(x, y, attrBackground)
	}
}

func (d *Display) clearRow(startx, y int, bg termbox.Attribute) {
	for x := startx; x < d.size.x; x++ {
		termbox.SetCell(x, y, ' ', colorBlack, bg)
	}
}

func (d *Display) writeText(str string, x, y int, fg, bg termbox.Attribute) int {
	runes := []rune(str)
	for _, c := range runes {
		if x >= d.size.x {
			break
		}
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
	return x
}

func (d *Display) Resize(x, y int) {
	d.lock()
	defer d.unlock()

	d.size.x = x
	d.size.y = y

	// reset termbox
	termbox.Clear(colorBlack, colorBlack)

	// draw everything
	d.drawBasics()
	d.drawChannel()
	d.drawTrackTitle()
	d.drawPlaying()
	d.drawTime()
	d.drawChannelList()
	//++ TODO: re-write current notification; move notification drawing to drawNotification()

	// flush to terminal
	termbox.Flush()
}

func (d *Display) Notify(message string) {
	d.chNotify <- message
}

func (d *Display) SetChannel(channelName, channelKey string) {
	d.lock()
	defer d.unlock()
	d.channelName = channelName
	d.channelKey = channelKey
	d.drawChannel()
	termbox.Flush()
}

func (d *Display) SetTrackTitle(title string) {
	d.lock()
	defer d.unlock()
	d.trackTitle = title
	d.drawTrackTitle()
	termbox.Flush()
}

func (d *Display) SetPlaying(playing bool) {
	d.lock()
	defer d.unlock()
	d.playing = playing
	d.drawPlaying()
	d.drawChannelList()
	termbox.Flush()
}

func (d *Display) SetTrackDuration(duration, passed time.Duration) {
	d.lock()
	defer d.unlock()
	d.trackDuration = duration
	d.trackPassed = passed
	d.drawTime()
	termbox.Flush()
}

func (d *Display) SetVolume(volume int) {
	d.lock()
	defer d.unlock()
	d.volume = volume
	d.drawVolume()
	termbox.Flush()
}

func (d *Display) SetChannelList(chl []*channelInfo) {
	d.lock()
	defer d.unlock()
	d.channelList = chl
	d.drawChannelList()
	termbox.Flush()
}

func (d *Display) MoveChannelListSelection(m int) {
	d.lock()
	defer d.unlock()
	d.channelListSelected += m
	if d.channelListSelected < 0 {
		d.channelListSelected = 0
	} else if d.channelListSelected >= len(d.channelList) {
		d.channelListSelected = len(d.channelList) - 1
	}
	d.drawChannelList()
	termbox.Flush()
}

func (d *Display) GetChannelSelection() string {
	return d.channelList[d.channelListSelected].channelKey
}
