package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// _ "net/http/pprof"

	"github.com/PuerkitoBio/goquery"
)

var dryrun = flag.Bool("dry", false, "don't download anything")
var verbose = flag.Bool("v", false, "verbose")
var verbose2 = flag.Bool("vv", false, "double verbose (prints more things)")
var verbose3 = flag.Bool("vvv", false, "triple verbose (prints even more things)")
var threads = flag.Int("threads", 1, "number of threads")

var urlstodo = make(chan string, 99999)
var done = make(chan bool, 1)

func main() {
	flag.Parse()
	if *verbose2 {
		*verbose = *verbose2
	}
	if *verbose3 {
		*verbose2 = *verbose3
		*verbose = *verbose3
	}

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	allurls := flag.Args()
	if *threads <= 1 {
		for i := 0; i < len(allurls); i++ {
			loc, err := url.Parse(allurls[i])
			if err != nil {
				log.Print(err)
				return
			}
			for _, v2 := range DownloadOrGetLinks(allurls[i]) {
				loc2, err := url.Parse(v2)
				if err != nil {
					log.Print(err)
					return
				}
				if loc.Hostname() != loc2.Hostname() || len(loc2.Path) <= len(loc.Path) {
					if *verbose2 {
						fmt.Println("skipping", v2)
					}
				} else {
					allurls = append(allurls, v2)
				}
			}
		}
	} else {
		for _, v := range allurls {
			urlstodo <- v
		}

		for i := 0; i < *threads; i++ {
			go func() {
				for {
					select {
					case res := <-urlstodo:
						DoURL(res)
					case <-time.After(20 * time.Second):
						done <- true
						return
					}
				}
			}()
		}
		for i := 0; i < *threads; i++ {
			<-done
		}
	}
}

func DoURL(urlx string) {
	loc, err := url.Parse(urlx)
	if err != nil {
		log.Print(err)
		return
	}
	for _, v2 := range DownloadOrGetLinks(urlx) {
		loc2, err := url.Parse(v2)
		if err != nil {
			log.Print(err)
			return
		}
		if loc.Hostname() != loc2.Hostname() || len(loc2.Path) <= len(loc.Path) {
			if *verbose2 {
				fmt.Println("skipping", v2)
			}
		} else {
			// allurls = append(allurls, v2)
			urlstodo <- v2
		}
	}
}

func DownloadOrGetLinks(urlx string) (urls []string) {
	resp, err := http.Get(urlx)
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("status code error: %d %s", resp.StatusCode, resp.Status)
		return
	}
	if strings.Index(resp.Header.Get("Content-Type"), "text/html") >= 0 {
		if *verbose2 {
			fmt.Println("checking", urlx)
		}
		return GetAllLinksFromResp(urlx, resp)
	} else if !*dryrun {
		loc, err := url.Parse(urlx)
		if err != nil {
			log.Print(err)
			return
		}
		pathslice := strings.Split(loc.Path, "/")
		if len(pathslice) > 0 {
			filename := pathslice[len(pathslice)-1]
			path := pathslice[:len(pathslice)-1]
			if len(path) > 0 && path[0] == "" {
				path = path[1:]
			}
			err = os.MkdirAll(strings.Join(path, "/"), os.ModePerm)
			if err != nil {
				log.Print(err)
				return
			}
			path = append(path, filename)
			if _, err := os.Stat(strings.Join(path, "/")); os.IsNotExist(err) {
				if *verbose {
					fmt.Println("downloading", urlx)
				}
				f, err := os.Create(strings.Join(path, "/"))
				if err != nil {
					log.Print(err)
					return
				}
				defer f.Close()

				io.Copy(f, resp.Body)
			}
		}
	}
	return
}

func GetAllLinksFromResp(urlx string, resp *http.Response) (urls []string) {
	loc, err := url.Parse(urlx)
	if loc2, err2 := resp.Location(); err2 == nil {
		loc = loc2
	} else if err != nil {
		log.Print(err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("href")
		if exists {
			if *verbose3 {
				fmt.Println("found", val)
			}
			loc2, err := url.Parse(val)
			if err != nil {
				log.Print(err)
				return //continue
			}
			if loc2.Scheme == "" {
				loc2.Scheme = loc.Scheme
			}
			if loc2.Host == "" {
				loc2.Host = loc.Host
			}
			if len(loc2.Path) > 0 && loc2.Path[0] != '/' {
				urltmparr := strings.Split(loc.Path, "/")
				if len(urltmparr) > 0 {
					if len(loc2.Path) > 2 && loc2.Path[:2] == "./" {
						loc2.Path = loc2.Path[2:]
					}
					urltmparr[len(urltmparr)-1] = loc2.Path
					loc2.Path = strings.Join(urltmparr, "/")
				} else {
					loc2.Path = "/" + loc2.Path
				}
			}

			urls = append(urls, loc2.String())
		}
	})
	return
}
