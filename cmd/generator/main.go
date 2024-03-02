package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	i := 0

	for {

		f, err := os.Create(fmt.Sprintf("./tmp/test%d.txt", i))

		if err != nil {
			log.Fatalf("Error creating file: %s", err)
			break
		}
		defer f.Close()

		str := strings.Repeat("Hello world", 1024*1024)

		f.WriteString(str)

		i++

		if i == 10 {
			break
		}
	}

}
