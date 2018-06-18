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
var matches []Matches

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

type AllMatchesPage struct {
	Title string
	Match []Matches
}

type HomePage struct {
	Title string
	Links map[string]string
}

func allmatches(c chan []Matches) {
	defer wg.Done()
	resp, err := http.Get("http://worldcup.sfg.io/matches")
	if err != nil {
		fmt.Println("No json for you")
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &matches)
	c <- matches
}

func fifaMatches(w http.ResponseWriter, r *http.Request) {
	queue := make(chan []Matches, 50)
	wg.Add(1)
	go allmatches(queue)
	wg.Wait()
	close(queue)
	page := AllMatchesPage{Title: "FIFA WORLDCUP 2K18 MATCHES", Match: matches}
	t, _ := template.ParseFiles("matches.html")
	t.Execute(w, page)
}

func todayMatches(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://worldcup.sfg.io/matches/today")
	if err != nil {
		fmt.Println("No json for you")
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &matches)
	page := AllMatchesPage{Title: "FIFA WORLDCUP 2K18 TODAY'S BATTLES", Match: matches}
	t, _ := template.ParseFiles("matches.html")
	t.Execute(w, page)
}

func fifa(w http.ResponseWriter, r *http.Request) {
	links := make(map[string]string)
	links["/matches"] = "FIFA WORLDCUP 2K18 ALL MACTHES"
	links["/matches/today"] = "ALL MACTHES TO BE PLAYED TODAY"
	page := HomePage{Title: "FIFA WORLDCUP 2K18", Links: links}
	t, _ := template.ParseFiles("home.html")
	t.Execute(w, page)
}

func main() {
	server := http.Server{
		Addr: ":" + os.Getenv("PORT"), //
	}
	http.HandleFunc("/", fifa)
	http.HandleFunc("/matches", fifaMatches)
	http.HandleFunc("/matches/today", todayMatches)
	//http.ListenAndServe("localhost:8000", nil)
	server.ListenAndServe()
}
