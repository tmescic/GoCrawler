package main

import (
	"fmt"
	"bufio"
	"os"
	"net/http"
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
	
		fmt.Println("Body       :", resp.Body)
		fmt.Println()
	}

	err3 := scanner.Err()
	check(err3)

	
	fmt.Println("Not a crawler yet, but I'm getting there...")
}

