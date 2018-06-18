package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var wg sync.WaitGroup

type Matches struct {
	Venue      string   `json:"venue"`
	Location   string   `json:"location"`
	Datetiming string   `json:"datetime"`
	Status     string   `json:"status"`
	HomeTeam   HomeTeam `json:"home_team"`
	AwayTeam   AwayTeam `json:"away_team"`
	Winner     string   `json:"winner"`
	WinnerCode string   `json:"winner_code"`
}

type HomeTeam struct {
	Country string `json:"country"`
	Code    string `json:"code"`
	Goals   int    `json:"goals"`
}

type AwayTeam struct {
	Country string `json:"country"`
	Code    string `json:"code"`
	Goals   int    `json:"goals"`
}

type FifaPage struct {
	Title string
	Match []Matches
}

var matches []Matches

func allmatches(c chan []Matches) {
	defer wg.Done()
	resp, err := http.Get("http://worldcup.sfg.io/matches")
	if err != nil {
		fmt.Println("No json for you")
	}
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(bytes, &matches)
	c <- matches
}

func fifaMatches(w http.ResponseWriter, r *http.Request) {
	queue := make(chan []Matches, 50)
	wg.Add(1)
	go allmatches(queue)
	wg.Wait()
	close(queue)
	page := FifaPage{Title: "FIFA WORLDCUP 2K18 MATCHES", Match: matches}
	t, _ := template.ParseFiles("fifa.html")
	t.Execute(w, page)
}

func fifa(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<h1>Details of Every FIFA WORLDCUP 2K18 MATCHES</h1>
		<a href="/matches" target='_blank'>All Matches</a>
		`)
}

func main() {
	server := http.Server{
		Addr: ":" + os.Getenv("PORT"), //
	}
	http.HandleFunc("/", fifa)
	http.HandleFunc("/matches", fifaMatches)
	//http.ListenAndServe("localhost:8000", nil)
	server.ListenAndServe()
}
