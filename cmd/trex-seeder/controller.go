package main

import (
	"log"
	"time"

	"github.com/ziscky/Taipei-Torrent/torrent"
	"github.com/ziscky/trex/common"
)

//Controller controls the operation,starting and running of the initial seeders
//seeders -> array of seed workers
//listener -> listener for the torrent registry file
//torrents -> torrents being seeded by the seed workers
//options -> various configurable options for the torrent sessions
type Controller struct {
	seeders  []*common.Seeder
	listener *common.Listener
	torrents []string
	stop     chan struct{}
	options  *torrent.TorrentFlags
}

//NewController creates a controller instance
func NewController(seeders []*common.Seeder, listener *common.Listener, torrents []string, flags *torrent.TorrentFlags, stop chan struct{}) *Controller {
	return &Controller{
		seeders:  seeders,
		listener: listener,
		torrents: torrents,
		stop:     stop,
		options:  flags,
	}
}

//Listen listens for changes in the torrent registry
func (c *Controller) Listen() {
	signal := make(chan struct{})
	go c.listener.Start(signal)

	for {
		select {
		case <-signal:
			log.Println("signaled")
			c.RedistributeWork()
		}
	}
}

//RedistributeWork reloads the seeders with new torrents if any
func (c *Controller) RedistributeWork() {
	c.torrents = checkTorrents(*registry)
	c.DistributeWork()
	c.StopSeeders()
	c.StartSeeders()
}

//DistributeWork distributes the torrents to the seeders
func (c *Controller) DistributeWork() {
	//clear seeder work queues
	for _, s := range c.seeders {
		if s.Running {
			s.Stop()
		}
		s.Clear()
	}

	x := 0
	for _, work := range c.torrents {
		c.seeders[x].AddWork(work)
		x++
		if x == *numSeeders {
			x = 0 //reset
		}
	}
}

//StartSeeders starts the seeders
func (c *Controller) StartSeeders() {
	for _, seeder := range c.seeders {
		go seeder.Start(c.options)
		time.Sleep(10 * time.Millisecond) //pause before binding on ports
	}
}

//StopSeeders stops the running seeders
func (c *Controller) StopSeeders() {
	for _, seeder := range c.seeders {
		if seeder.Running {
			seeder.Stop()
		}
	}
}
