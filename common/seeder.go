package common

import (
	"log"

	"github.com/ziscky/Taipei-Torrent/torrent"
)

//Seeder represents a seed worker
//ID -> seeder ID
// NumTasks -> number of tasks the seeder has
//Running -> flag to check whether seeder is running
//work -> the work being processed by the seed worker
//stop -> channel to stop the seeder
type Seeder struct {
	ID       int
	NumTasks int
	Running  bool
	work     []string
	stop     chan struct{}
}

//NewSeeder returns a seeder instance
func NewSeeder(id int) *Seeder {
	return &Seeder{
		ID:   id,
		stop: make(chan struct{}),
	}
}

//AddWork adds work for the seed worker
func (seeder *Seeder) AddWork(work string) {
	seeder.work = append(seeder.work, work)
}

//Start starts the torrent session for the seed worker
func (seeder *Seeder) Start(flags *torrent.TorrentFlags) {
	log.Println(seeder.ID, "started")
	flags.Port++
	seeder.Running = true
	defer func() { //flag seeder as stopped incase of a crash
		seeder.Running = false
	}()

	err := torrent.RunTorrents(flags, seeder.work, seeder.stop)
	if err != nil {
		log.Fatal("Could not run torrents", seeder.work, err)
	}
}

//Stop stops the seed worker
func (seeder *Seeder) Stop() {
	seeder.stop <- struct{}{}
	seeder.Running = false
}

//Status gets all info about the seeder
func (seeder *Seeder) Status() []string {
	return seeder.work
}

//Clear clears a seed workers work load
func (seeder *Seeder) Clear() {
	seeder.work = []string{}
}
