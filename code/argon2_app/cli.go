package main

import (
    "fmt"
    //"log"
    "net/http"
    "io"
    "os"
    "strings"
)

func main() {
    password := "DEFAULT_PASSWORD"
    if len(os.Args) >= 2 {
        password = os.Args[1]
    }
    url := fmt.Sprintf("http://localhost:3333/?password=%s", password)
    fmt.Printf("URL: %s\n", url)
   	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	body, _ := io.ReadAll(res.Body)
    data := strings.TrimSpace(string(body))

	fmt.Printf("client: status: %d, %s\n", res.StatusCode, data)
}

