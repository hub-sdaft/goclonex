package main

import (
	"fmt"

	"github.com/hub-sdaft/goclonex/pkg/memos"
)

func main() {
	os := memos.NewInMemoryOS("C:/Users/user/Documents")

	err := os.MkdirAll("/a/b/c", 0)
	if err != nil {
		fmt.Printf("err1: %s", err)
		return
	}

	fmt.Printf("\n\n%#v\n\n", os.Root)

	f, err := os.Create("/a/b/c/file.txt")
	if err != nil {
		fmt.Printf("err2: %s", err)
		return
	}

	fmt.Printf("\n\n%+v\n\n", os.Root)

	fmt.Printf("%v", f)
}