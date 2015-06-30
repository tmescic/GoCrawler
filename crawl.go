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

	file, err := os.Open("seeds.txt")
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentUrl := scanner.Text()
		fmt.Println("Visiting   :", currentUrl)
		resp, err2 := http.Get(currentUrl)
		check(err2)

		fmt.Println("StatusCode :", resp.StatusCode)

		links := parsePage(resp.Body)

		fmt.Println("Found links:", len(links))
		parsePage(resp.Body)
		fmt.Println()
	}

	err3 := scanner.Err()
	check(err3)

	fmt.Println("Not a crawler yet, but I'm getting there...")
}

func parsePage(body io.ReadCloser) (links [2]string) {

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
