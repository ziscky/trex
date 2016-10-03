package main

import (
	"bufio"
	"flag"
	"fmt"

	"log"
	"math"

	"os"

	"os/signal"

	"github.com/ziscky/Taipei-Torrent/torrent"
	"github.com/ziscky/trex/common"
	"golang.org/x/net/proxy"
)

var (
	torrentPath         = flag.String("path", "torrent", "Path to the torrents to track. --path=/path/to/torrents")
	numSeeders          = flag.Int("numSeeds", 20, "Number of concurrent seeders to deploy")
	numSession          = flag.Int("numSessions", 10, "Number of torrents each seeder seeds")
	port                = flag.Int("port", 18000, "Start port to be used for seeders,increases +1 for each seeder ")
	verbose             = flag.Bool("verbose", false, "Be verbose")
	fileDir             = flag.String("fileDir", "content", "path to directory where files are stored")
	registry            = flag.String("registry", "torrent-reg", "Path to where torrents are registered")
	useDeadlockDetector = flag.Bool("useDeadlockDetector", false, "Panic and print stack dumps when the program is stuck.")
	useLPD              = flag.Bool("useLPD", false, "Use Local Peer Discovery")
	useUPnP             = flag.Bool("useUPnP", false, "Use UPnP to open port in firewall.")
	useNATPMP           = flag.Bool("useNATPMP", false, "Use NAT-PMP to open port in firewall.")
	gateway             = flag.String("gateway", "", "IP Address of gateway.")
	useDHT              = flag.Bool("useDHT", false, "Use DHT to get peers.")
	trackerlessMode     = flag.Bool("trackerlessMode", false, "Do not get peers from the tracker. Good for testing DHT mode.")
	proxyAddress        = flag.String("proxyAddress", "", "Address of a SOCKS5 proxy to use.")
	initialCheck        = flag.Bool("initialCheck", true, "Do an initial hash check on files when adding torrents.")
	useSFTP             = flag.String("useSFTP", "", "SFTP connection string, to store torrents over SFTP. e.g. 'username:password@192.168.1.25:22/path/'")
	useRamCache         = flag.Int("useRamCache", 0, "Size in MiB of cache in ram, to reduce traffic on torrent storage.")
	useHdCache          = flag.Int("useHdCache", 0, "Size in MiB of cache in OS temp directory, to reduce traffic on torrent storage.")
	execOnSeeding       = flag.String("execOnSeeding", "", "Command to execute when torrent has fully downloaded and has begun seeding.")
	maxActive           = flag.Int("maxActive", 16, "How many torrents should be active at a time. Torrents added beyond this value are queued.")
	memoryPerTorrent    = flag.Int("memoryPerTorrent", -1, "Maximum memory (in MiB) per torrent used for Active Pieces. 0 means minimum. -1 (default) means unlimited.")

	seedRatio   = flag.Float64("seedRatio", math.Inf(0), "Seed until ratio >= this value before quitting.")
	quickResume = flag.Bool("quickResume", false, "Save torrenting data to resume faster. '-initialCheck' should be set to false, to prevent hash check on resume.")
)

func parseTorrentFlags() (flags *torrent.TorrentFlags, err error) {
	dialer, err := dialerFromFlags()
	if err != nil {
		return
	}
	flags = &torrent.TorrentFlags{
		Dial:                dialer,
		Port:                *port,
		FileDir:             *fileDir,
		SeedRatio:           *seedRatio,
		UseDeadlockDetector: *useDeadlockDetector,
		UseLPD:              *useLPD,
		UseDHT:              *useDHT,
		UseUPnP:             *useUPnP,
		UseNATPMP:           *useNATPMP,
		TrackerlessMode:     *trackerlessMode,
		Gateway:             *gateway,
		InitialCheck:        *initialCheck,
		FileSystemProvider:  fsproviderFromFlags(),
		Cacher:              cacheproviderFromFlags(),
		ExecOnSeeding:       *execOnSeeding,
		QuickResume:         *quickResume,
		MaxActive:           *maxActive,
		MemoryPerTorrent:    *memoryPerTorrent,
	}
	return
}

func cacheproviderFromFlags() torrent.CacheProvider {
	if (*useRamCache) > 0 && (*useHdCache) > 0 {
		log.Panicln("Only one cache at a time, please.")
	}

	if (*useRamCache) > 0 {
		return torrent.NewRamCacheProvider(*useRamCache)
	}

	if (*useHdCache) > 0 {
		return torrent.NewHdCacheProvider(*useHdCache)
	}
	return nil
}

func fsproviderFromFlags() torrent.FsProvider {
	if len(*useSFTP) > 0 {
		return torrent.NewSftpFsProvider(*useSFTP)
	}
	return torrent.OsFsProvider{}
}

func dialerFromFlags() (proxy.Dialer, error) {
	if len(*proxyAddress) > 0 {
		return proxy.SOCKS5("tcp", string(*proxyAddress), nil, &proxy.Direct)
	}
	return proxy.FromEnvironment(), nil
}

func checkTorrents(torrentRegistry string) []string {
	fileList := []string{}
	if finfo, err := os.Stat(torrentRegistry); err != nil {
		log.Fatal(torrentRegistry, " does not exist")
	} else {
		if finfo.IsDir() {
			log.Fatal(torrentRegistry, " can't be a directory")
		} else {
			file, err := os.Open(torrentRegistry)
			if err != nil {
				log.Fatal(err)
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if len(scanner.Text()) == 0 {
					continue
				}
				if _, err := os.Stat(scanner.Text()); err != nil {
					log.Println("Couldn't find ", scanner.Text())
					continue
				}
				fileList = append(fileList, scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			file.Close()

		}
	}
	return fileList
}

func main() {
	flag.Parse()
	log.SetFlags(log.Llongfile | log.LstdFlags)
	torrentFlags, err := parseTorrentFlags()
	if err != nil {
		log.Fatal("Could not parse flags:", err)
	}
	torrents := checkTorrents(*registry)

	if *verbose {
		fmt.Println("Watching:", len(torrents))
		for _, t := range torrents {
			fmt.Println("\t", t)
		}
	}

	var seeders []*common.Seeder

	for i := 0; i < *numSeeders; i++ {
		seeders = append(seeders, common.NewSeeder(i))
	}

	//start .torrent listener
	stopListener := make(chan struct{})
	listener, err := common.NewListener(3, *registry, stopListener)
	if err != nil {
		log.Fatal(err)
	}

	app := NewController(seeders, listener, torrents, torrentFlags, stopListener)
	go app.Listen()

	app.DistributeWork()
	app.StartSeeders()

	log.Println("Waiting")
	quit := listenSigInt()
	<-quit
	log.Println("Waiting2")
	//stop folder listener

	//stop seeders
	fmt.Println("Stopping seeders")
	app.StopSeeders()
	// listener.Stop <- struct{}{}

	fmt.Println("Goodbye")

	return

}

func listenSigInt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}
