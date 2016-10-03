package main

import (
	"bufio"
	"fmt"
	"log"
	"path"

	"os"

	"path/filepath"

	"github.com/jackpal/Taipei-Torrent/torrent"

	"flag"

	"github.com/ziscky/trex/common"
)

var (
	contentPath = flag.String("content", "content", "Path to the properly formatted content")
	torrentPath = flag.String("torrent", "torrent", "Path to where to store torrents")
	registry    = flag.String("registry", "torrent-reg", "Path to file where all torrents are registered")
	addr        = flag.String("addr", "tracker.pdftrex.org", "Address of the torrent primary tracker")
)

//TODO: check if contents of torrent-reg are valid
func checkContent(contentPath string) []string {
	torrentExists := func(name string) bool {
		file, err := os.Open(*registry)
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if len(scanner.Text()) == 0 {
				continue
			}
			p, _ := filepath.Abs(path.Join(*torrentPath, name))
			if p == scanner.Text() {
				return true
			}

		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		file.Close()
		return false
	}

	fileList := []string{}

	if _, err := os.Stat(contentPath); err != nil {
		log.Fatal(contentPath, " does not exist")
	}
	if err := filepath.Walk(contentPath, func(path string, f os.FileInfo, err error) error {

		//ignore base path
		if path == contentPath {
			return nil
		}

		//check if a torrent file for the content is created
		if !torrentExists(common.NameHash(f.Name())) {
			createTorrent(path, *torrentPath+"/"+common.NameHash(f.Name())+".torrent")
		}
		fileList = append(fileList, path)
		return nil
	}); err != nil {
		log.Fatal("Couldn't read ", contentPath)
	}

	return fileList
}
func printErrors(x map[string][]string) {
	fmt.Println("------------------Error Summary----------------")
	for k, v := range x {
		fmt.Println(k, ":")
		for _, j := range v {
			fmt.Println("\t", j)
		}

	}
}

func createTorrent(path string, outfile string) {
	log.Println(os.Getwd())
	f, err := os.OpenFile(outfile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal("Could not open torrent file:", err)
	}
	defer f.Close()
	err = torrent.WriteMetaInfoBytes(path, *addr, f)
	if err != nil {
		log.Fatal("Could not create torrent file:", err)
	}
	return
}
func checkTorrents(torrentPath string) []string {
	fileList := []string{}
	if finfo, err := os.Stat(torrentPath); err != nil {
		log.Fatal("Path does not exist")
	} else {
		if !finfo.IsDir() {
			log.Fatal("Path is not a directory")
		} else {
			if err := filepath.Walk(torrentPath, func(path string, f os.FileInfo, err error) error {
				if f.IsDir() {
					return nil
				}
				abs, _ := filepath.Abs(path)
				fileList = append(fileList, abs)
				return nil
			}); err != nil {
				log.Fatal(err)
			}
		}
	}
	return fileList
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	content := checkContent(*contentPath)
	if len(content) == 0 {
		log.Println("No (new) content detected")
		os.Exit(1)
	}
	torrents := checkTorrents(*torrentPath)
	fmt.Println("Generated:")
	f, err := os.Create(*registry)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	for _, t := range torrents {
		fmt.Println("\t", t)
		f.WriteString(t + "\n")
	}

}
