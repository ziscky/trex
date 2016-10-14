package common

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

//Listener holds info relevant to monitoring a file for changes over
//a specified time
type Listener struct {
	Interval int
	File     string
	LastHash string
	Stop     chan struct{}
}

//NewListener checks whether the given file to watch is valid and returns a listener instance
func NewListener(interval int, registry string, stop chan struct{}) (*Listener, error) {
	_, err := os.Stat(registry)
	if err != nil {
		return nil, err
	}
	return &Listener{
		File:     registry,
		Interval: interval,
	}, nil
}

//Hash gets an md5 hashsum of the monitored file
func (listener *Listener) Hash() string {
	hash := md5.New()
	finfo, err := os.Stat(listener.File)
	if err != nil {
		log.Fatal("Listener not active, restart program:", err)
	}
	io.WriteString(hash, listener.File)
	fmt.Fprintf(hash, "%v", finfo.Size())
	fmt.Fprintf(hash, "%v", finfo.ModTime())
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//Start begins polling the file for the md5 hash and checking for changes
//changes reported to the passed signal channel
func (listener *Listener) Start(signal chan struct{}) {
	//initial hash
	listener.LastHash = listener.Hash()
	log.Println((listener.LastHash))
	timer := time.NewTicker(time.Duration(listener.Interval) * time.Second)
	for {
		select {
		case <-timer.C:
			curHash := listener.Hash()
			if curHash != listener.LastHash {
				signal <- struct{}{}
			}
			listener.LastHash = curHash

		case <-listener.Stop:
			close(signal)
			return
		}
	}
}
