package main

import (
	"fmt"
	"time"
)

// Scratch main faile to expriment short code snippets.

func main() {
	fmt.Println(time.Now().UTC().Format(time.UnixDate))
}
