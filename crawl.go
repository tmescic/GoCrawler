package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
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

	// visited nodes are stored in this maps
	visited := make(map[string]bool)

	// queue of pages to be visited
	queue := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentUrl := scanner.Text()
		fmt.Println("Visiting   :", currentUrl)
		resp, err2 := http.Get(currentUrl)
		check(err2)

		fmt.Println("StatusCode :", resp.StatusCode)

		parsePage(resp.Body, queue)
		visited[currentUrl] = true;
	}

	err3 := scanner.Err()
	check(err3)

	fmt.Println("Not a crawler yet, but I'm getting there...")
	fmt.Println("Queue size: ", len(queue))
	fmt.Println("Visited : ", len(visited))
	
}

func parsePage(body io.ReadCloser, queue []string) {

	doc, err := html.Parse(body)
	if err != nil {
		fmt.Println("Error : ", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for i := 0; i < len(n.Attr); i++ {
				if n.Attr[i].Key == "href" {
					fmt.Println("new link: ", n.Attr[i].Val)
					queue = append(queue, n.Attr[i].Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return
}
