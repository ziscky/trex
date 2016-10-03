package main

import (
	"log"
	"path"

	"github.com/ziscky/Taipei-Torrent/torrent"
	"github.com/ziscky/Taipei-Torrent/tracker"
	"github.com/ziscky/trex/common"
	"golang.org/x/net/proxy"
)

//Controller controls the configuration,starting stopping of the tracker
//Torrents -> array of paths to torrent metainfo files to be supported by the tracker
//listener -> listens for changes in the torrent registry
//dialer -> net dial through proxy
type Controller struct {
	Tracker  *tracker.Tracker
	Torrents []string
	listener *common.Listener
	dialer   proxy.Dialer
}

//NewController helper to create a controller object
func NewController(tracker *tracker.Tracker, listener *common.Listener, dialer proxy.Dialer, torrents []string) *Controller {
	return &Controller{
		Tracker:  tracker,
		listener: listener,
		Torrents: torrents,
		dialer:   dialer,
	}
}

//StartTracker starts the tracker
func (c *Controller) StartTracker() error {
	for _, torrentFile := range c.Torrents {
		var metaInfo *torrent.MetaInfo
		metaInfo, err := torrent.GetMetaInfo(c.dialer, torrentFile)
		if err != nil {
			continue
		}
		name := metaInfo.Info.Name
		if name == "" {
			name = path.Base(torrentFile)
		}
		err = c.Tracker.Register(metaInfo.InfoHash, name)
		if err != nil {
			continue
		}
	}
	return c.Tracker.ListenAndServe()
}

//Reload checks the torrent registry for any new torrents and begins to support them
func (c *Controller) Reload() {
	c.Torrents = checkTorrents()

	for _, t := range c.Torrents {
		meta, err := torrent.GetMetaInfo(c.dialer, t)
		if err != nil {
			log.Println(err)
			continue
		}
		name := meta.Info.Name
		if name == "" {
			name = path.Base(t)
		}
		if !c.Tracker.Exists(meta.InfoHash) {
			if err := c.Tracker.Register(meta.InfoHash, name); err != nil {
				log.Println("Couldn't register ", meta.Info.Name, err)
				continue
			}

		}
	}

}

//Listen listens for changes to the torrent registry
func (c *Controller) Listen() {
	signal := make(chan struct{})
	go c.listener.Start(signal)

	for {
		select {
		case <-signal:
			c.Reload()
		}
	}
}

//Stop stops the tracker
func (c *Controller) Stop() error {
	return c.Tracker.Quit()
}
