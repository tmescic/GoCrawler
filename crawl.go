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
	"runtime"
	"regexp"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}


type IgnoreRule struct {
	rule *regexp.Regexp
	inverse  bool
} 
	
func main() {

	startTime := time.Now().UnixNano() / 1e6

	// visited nodes are stored in this maps
	visited := make(map[string]bool)

	// dead links
	dead := make(map[string]bool)

	// queue of pages to be visited
	queue := make(map[string]bool)

	// Loads the ignore list. Every new url is checked against the patterns in 
	// the ignore list. If it matches one of them, then it is not added to the queue.
	ignoresTxt := getLinesFromFile("ignore.txt") 
	ignores := make([]IgnoreRule, len(ignoresTxt))
	for i, pattern := range ignoresTxt {
		inv := strings.HasPrefix(pattern, "!")
		if inv {
			pattern = pattern[1:] // remove the "!"
			fmt.Println("new pattern : ", pattern)
		}
		compiledRule := regexp.MustCompile(pattern)
		ignores[i] = IgnoreRule{rule : compiledRule, inverse : inv} 
	}

	// handle Ctrl+C (print statistics)
	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		
		printStats(startTime, visited, dead, queue)

		os.Exit(0)
	}()

	seeds := getLinesFromFile("seeds.txt")
	for _, seed := range seeds {
		queue[seed] = true
	}

	// visit pages from the queue until it's empty (that will never happen)
	for len(queue) > 0 {
		currentUrl := getNext(queue)
		delete(queue, currentUrl)
		oldQueueLen := len(queue)
		fmt.Print(len(visited), ": ", currentUrl)
		
		resp, err2 := http.Get(currentUrl)

		if (err2 == nil) {

			fmt.Print(" [R:", resp.StatusCode)
			links := parsePage(resp.Body)
			visited[currentUrl] = true
			fmt.Print(", F:", len(links))

			// add new links to queue
			add(queue, links, currentUrl, visited, ignores)

			fmt.Print(", N:",(len(queue) - oldQueueLen),", Q:",len(queue),"]\n")
		} else {
			fmt.Println("\nError while fetching:", currentUrl)
			dead[currentUrl] = true
		}
	}
	
	printStats(startTime, visited, dead, queue)
}

func printStats(startTime int64, visited map[string]bool, dead map[string]bool, queue map[string]bool) {

	d := time.Now().UnixNano()/1e6 - startTime

	fmt.Println("\n\nStatistics")
	fmt.Println("==========\n")
	fmt.Println("Logical cores    :", runtime.NumCPU())
	fmt.Println("Duration         :", d, "ms")
	fmt.Println("Visited          :", len(visited))
	fmt.Println("Dead links       :", len(dead))
	fmt.Println("Still in queue   :", len(queue))

	fmt.Printf("Pages per second : %f\n", (float64(len(visited)) / float64(d) * 1000.))
}

// Opens the file with given name, reads it, and returns an array all 
// lines. If a line starts with # it is ignored (comment).
func getLinesFromFile(fileName string) []string {

	lines := make ([]string, 0)

	file, err := os.Open(fileName)
	check(err)
	defer file.Close()

	// read lines to an array
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		newLine := scanner.Text()
		if !strings.HasPrefix(newLine, "#") {
			lines = append(lines, scanner.Text())
		}
	}
	return lines
}

// Parses the given page and return all links found on the page.
// Duplicates are possible in the returned array.
func parsePage(body io.ReadCloser) []string {

	doc, err := html.Parse(body)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	var visit func(*html.Node) []string
	visit = func(n *html.Node) []string {
		links := make([]string, 0)

		if n.Type == html.ElementNode && n.Data == "a" {
			for i := 0; i < len(n.Attr); i++ {
				if n.Attr[i].Key == "href" {
					links = append(links, n.Attr[i].Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			links = append(links, visit(c)...)
		}
		return links
	}

	return visit(doc)

}

// Formats the raw link found in the page (e.g. ingores javascript links, adds hostname 
// to relative links etc.)
func formatUrl(link string, origPageUrl string) string {

	ret := ""

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

// Returns the next item from the queue. If the queue is empty, an empty string is returned.
// Item is selected from the queue randomly.
func getNext(queue map[string]bool) string {
	for k := range queue {
		return k
	}
	return ""
}

// Adds the new link to the queue. The link is formatted first. If it's already visited that 
// it's not added to the queue
func add(queue map[string]bool, newLinks []string, origPageUrl string, visited map[string]bool, ignores []IgnoreRule) {
	for _, link := range newLinks {
		formatted := formatUrl(link, origPageUrl)
		if formatted != "" && !visited[formatted] {
			ignoreIt := false
			for _, ignore := range ignores {
				if (!ignore.inverse && ignore.rule.MatchString(formatted)) || (ignore.inverse && !ignore.rule.MatchString(formatted)) {
					ignoreIt = true
					break
				}

			}
			if !ignoreIt {
				// Add it to queue if it's not on the ignore list
				queue[formatted] = true
			} else {
				// fmt.Println("Ignoring: ", formatted)
			}
		}
	}
}

