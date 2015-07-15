package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"io"
	"golang.org/x/net/html"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	// inital seeds
	file, err := os.Open("seeds.txt")
	check(err)
	defer file.Close()

	// queue of pages to be visited
	queue := make([]string, 0)

	// put all seeds to queue
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		queue = append(queue, url)
	}
	err3 := scanner.Err()
	check(err3)


	// visited nodes are stored in this maps
	visited := make(map[string]bool)

	// visit pages from the queue
	for len(queue) > 0 {
		url := queue[0];
		queue = queue[1:] 
		oldQueueLen := len(queue)
		fmt.Print(len(visited), ": ", url)
		resp, err2 := http.Get(url)
		check(err2)

		fmt.Print(", response: ", resp.StatusCode)
		fmt.Print(", parsing... ")
		
		links := parsePage(resp.Body)

		visited[url] = true;
		
		fmt.Print(" links total: ", len(links))

		// see if they are already visited
		for _,link := range links {
			
			formatted := formatUrl(link)		

			if formatted != "" && !visited[formatted] {
				// TODO handle same link on a single page
				queue = append(queue, formatted)
			}
		}

		fmt.Print(", new links: ", (len(queue) - oldQueueLen))
		fmt.Println(", in queue: ", len(queue))
	}

	fmt.Println("Not a crawler yet, but I'm getting there...")
	fmt.Println("Queue size: ", len(queue))
	fmt.Println("Visited : ", len(visited))
	
}

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
					// fmt.Println("new link: ", n.Attr[i].Val)
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

func formatUrl(url string) string {

	var ret string	

	if url == "javascript:;" || url == "#" || url == "/" {
		ret = ""
	} else if strings.HasPrefix(url, "//") {
		ret = "http:" + url
	}

	return ret
}
