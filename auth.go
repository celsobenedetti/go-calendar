package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const (
	// TODO: configure this with viper
	credsFile = "/home/celso/.gocal/credentials.json"
	tokFile   = "/home/celso/.gocal/token.json"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authCodeCh := make(chan string)

	port := serveRandomPort(authCodeCh)
	config.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	go openBrowser(authURL)
	code := <-authCodeCh

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func openBrowser(url string) {
	err := exec.Command("xdg-open", url).Start()
	if err != nil {
		log.Fatalln(err)
	}
}

func serveRandomPort(authCodeCh chan string) int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		authCodeCh <- r.URL.Query().Get("code")
	})
	go http.Serve(listener, handler)

	return listener.Addr().(*net.TCPAddr).Port
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Remove a token to a file path.
func removeToken(path string) {
	err := os.Remove(path)
	// TODO: this is not fatal bruh
	if err != nil {
		log.Fatalf("Unable to remove oauth token file %s: %v", path, err)
	}
}

// If modifying these scopes, delete your previously saved token.json.
func readConfig() *oauth2.Config {
	b, err := os.ReadFile(credsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}
