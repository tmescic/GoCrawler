package main

import (
	"bufio"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	startTime := time.Now().UnixNano() / 1e6

	// visited nodes are stored in this maps
	visited := make(map[string]bool)

	// dead links
	dead := make(map[string]bool)

	// queue of pages to be visited
	queue := make([]string, 0)

	// handle Ctrl+C (print statistics)
	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		d := time.Now().UnixNano()/1e6 - startTime
		fmt.Println("\n\nStatistics")
		fmt.Println("==========\n")
		fmt.Println("Duration         :", d, "ms")
		fmt.Println("Visited          :", len(visited))
		fmt.Println("Dead links       :", len(dead))
		fmt.Println("Still in queue   :", len(queue))
		fmt.Printf("Pages per second : %f\n", (float64(len(visited)) / float64(d) * 1000.))

		os.Exit(0)
	}()

	// inital seeds
	file, err := os.Open("seeds.txt")
	check(err)
	defer file.Close()

	// put all seeds to queue
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seedUrl := scanner.Text()
		queue = append(queue, seedUrl)
	}
	err3 := scanner.Err()
	check(err3)

	// visit pages from the queue
	for len(queue) > 0 {
		currentUrl := queue[0]
		queue = queue[1:]
		oldQueueLen := len(queue)
		fmt.Print(len(visited), ": ", currentUrl)
		resp, err2 := http.Get(currentUrl)

		if (err2 == nil) {

			fmt.Print(", response:", resp.StatusCode, ", parsing...")
			links := parsePage(resp.Body)
			visited[currentUrl] = true
			fmt.Print(" links total: ", len(links))

			// see if they are already visited
			for link, _ := range links {

				formatted := formatUrl(link, currentUrl)

				if formatted != "" && !visited[formatted] {
					// TODO handle same link on a single page
					queue = append(queue, formatted)
				}
			}

			fmt.Println(", new links: ", (len(queue) - oldQueueLen), ", in queue: ", len(queue))
		} else {
			fmt.Println("\nError while fetching:", currentUrl)
			dead[currentUrl] = true
		}
	}
}

func parsePage(body io.ReadCloser) map[string]bool {

	doc, err := html.Parse(body)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	var visit func(*html.Node) map[string]bool
	visit = func(n *html.Node) map[string]bool {
		links := make(map[string]bool)

		if n.Type == html.ElementNode && n.Data == "a" {
			for i := 0; i < len(n.Attr); i++ {
				if n.Attr[i].Key == "href" {
					// fmt.Println("new link: ", n.Attr[i].Val)
					links[n.Attr[i].Val] = true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			newLinks := visit(c)
			for k, v := range newLinks {
				links[k] = v
			}
		}
		return links
	}

	return visit(doc)

}

func formatUrl(link string, origPageUrl string) string {

	var ret string

	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		ret = link
	} else if link == "javascript:;" || link == "#" || link == "/" {
		ret = ""
	} else if strings.HasPrefix(link, "//") { // missing http: at the beggining
		ret = "http:" + link
	} else if strings.HasPrefix(link, "/") { // relative link
		u, _ := url.Parse(origPageUrl)
		host := u.Host
		scheme := u.Scheme
		ret = scheme + "://" + host + link
	}
	return ret
}
