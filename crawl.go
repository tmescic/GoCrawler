package main

import (
	"fmt"
	"bufio"
	"os"
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
		fmt.Println(scanner.Text())
	}

	err2 := scanner.Err()
	check(err2)

	
	fmt.Println("Not a crawler yet, but I'm getting there...")
}

