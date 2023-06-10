package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	//"log"
	"net/http"
	//"io"
	"encoding/hex"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

type params struct {
	memory      uint32
	iterations  int
	parallelism uint8
	length      int
}

var ARGON_PARAMS = params{
	memory:      64 * 1024,
	iterations:  5000000,
	parallelism: 1,
	length:      16,
}

func main() {
	// TODO: imagining this will be a server
	http.HandleFunc("/", handleHttp)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

	/*
	   for {
	       token, err := handleRequest("password123")
	       if err != nil {
	           log.Fatal(err)
	       }

	       fmt.Println(token)
	   }
	*/
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("HTTP request: %s\n", r.URL)
	password := r.URL.Query()["password"][0]
	fmt.Printf("HTTP password: %s\n", password)

	token, _ := handleRequest(password)
	//fmt.Println(token)
	//io.WriteString(w, token)
	fmt.Fprintf(w, "Token: %s\n", hex.EncodeToString(token))
}

func handleRequest(text string) ([]byte, error) {
	var err error
	var hash, random, token []byte

	if hash, err = processInput(text); err != nil {
		return nil, err
	}
	if random, err = generateRandomBytes(ARGON_PARAMS.length); err != nil {
		return nil, err
	}

	token = make([]byte, ARGON_PARAMS.length)
	for i, v := range hash {
		token[i] = v ^ random[i]
	}

	return token, nil
}

func processInput(input string) ([]byte, error) {
	var p = ARGON_PARAMS
	var salt = make([]byte, p.length)
	//var hash = argon2.IDKey([]byte(input), salt, p.iterations, p.memory, p.parallelism, p.length)
	var hash = pbkdf2.Key([]byte(input), salt, p.iterations, p.length, sha256.New)

	return hash, nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}
