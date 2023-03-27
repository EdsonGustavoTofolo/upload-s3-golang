package main

import (
	"fmt"
	"os"
)

func main() {
	for i := 0; i < 20; i++ {
		f, err := os.Create(fmt.Sprintf("./temp/file%d.txt", i))
		if err != nil {
			panic(err)
		}
		_, err = f.WriteString("Hello GO lang world")
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
}
