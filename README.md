# Tune

Tune is a music-player for the [AudioAddict](https://www.audioaddict.com/) web radio's such as [Digitally Imported a.k.a. di.fm](https://di.fm) and [RadioTunes](https://radiotunes.com).

Currently tune is only available with a Command-line interface (cli). A Graphical User Interface (based on Qt) is on the roadmap.

## Features

Tune is still a work in progress. I've recently decided to open-source my work on it so far. Currently, features include:

- AudioAddict API implementation
- Player based on libvlc
- CLI interface
- Support for di.fm
- Persistent authentication and preferences

The following features are currently in progress or on the roadmap:

- Graphical user interface based on Qt/QML
- CI and versioned releases
- Compatibility with all AudioAddict networks
- Investigate more lightweight playback libraries
- Cross platform support (I'm just working with linux currently)

## Installing

Since releases aren't provided yet, you'll have to build manually. Make sure go is installed and run:

```sh
go get github.com/GeertJohan/tune/cmd/tune-cli
```

## History

I made an early version of tune in 2014 as a side project. It became more relevant to me when I uninstalled flash and found out Digitally Imported was still using it for the player on their website.
