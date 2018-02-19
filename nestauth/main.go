package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jsgoecke/nest"
)

var clientID = flag.String("client", "", "client id")
var state = flag.String("state", "STATE", "state")
var secret = flag.String("secret", "", "client secret")
var code = flag.String("code", "", "auth code")

func main() {
	flag.Parse()
	client := nest.New(*clientID, *state, *secret, *code)
	err := client.Authorize()
	if err != nil {
		log.Fatalf("error authorizing: %#v", err)
	}
	fmt.Printf("token: %s\n", client.Token)
	fmt.Printf("expires in: %d\n", client.ExpiresIn)
}
