package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
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

type AllTeams struct {
	ID          int    `json:"id"`
	Country     string `json:"country"`
	FifaCode    string `json:"fifa_code"`
	GroupID     int    `json:"group_id"`
	GroupLetter string `json:"group_letter"`
}

type Stats struct {
	ID          int `json:"id"`
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
	Losses      int `json:"losses"`
	Draws       int `json:"draws"`
	Points      int `json:"points"`
}

type AllMatchesPage struct {
	Title string
	Match []Matches
}

type AllTeamsPage struct {
	Title string
	Teams map[AllTeams]Stats
}

type HomePage struct {
	Title string
	Links map[string]string
}

func date() {
	for i := range matches {
		t, err := time.Parse(time.RFC3339, matches[i].Datetiming)
		if err != nil {
			fmt.Println(err)
		}
		timestr := t.String()
		pidx := strings.Index(timestr, "+")
		uidx := strings.Index(timestr, "U")
		matches[i].Datetiming = timestr[:pidx] + timestr[uidx:]
	}
}

func allmatches(c chan []Matches) {
	defer wg.Done()
	resp, err := http.Get("https://world-cup-json.herokuapp.com/matches")
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
	date()
	page := AllMatchesPage{Title: "FIFA WORLDCUP 2K18 MATCHES", Match: matches}
	t, _ := template.ParseFiles("matches.html")
	t.Execute(w, page)
}

func todayMatches(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://world-cup-json.herokuapp.com/matches/today")
	if err != nil {
		fmt.Println("No json for you")
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &matches)
	date()
	page := AllMatchesPage{Title: "FIFA WORLDCUP 2K18 TODAY'S BATTLES", Match: matches}
	t, _ := template.ParseFiles("matches.html")
	t.Execute(w, page)
}

func allteams(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://world-cup-json.herokuapp.com/teams/")
	if err != nil {
		fmt.Println("No json for you")
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	var teams []AllTeams
	json.Unmarshal(bytes, &teams)
	response, err := http.Get("https://world-cup-json.herokuapp.com/teams/results")
	if err != nil {
		fmt.Println("No json for you")
	}
	defer response.Body.Close()
	bites, _ := ioutil.ReadAll(response.Body)
	var stats []Stats
	json.Unmarshal(bites, &stats)
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].ID < teams[j].ID
	})
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].ID < stats[j].ID
	})
	teamdetails := make(map[AllTeams]Stats)
	for i := 0; i < len(teams); i++ {
		teamdetails[teams[i]] = stats[i]
	}
	page := AllTeamsPage{Title: "POINTS TABLE(GROUPWISE) FIFA WORLDCUP 2K18", Teams: teamdetails}
	t, _ := template.ParseFiles("teams.html")
	t.Execute(w, page)
}

func fifa(w http.ResponseWriter, r *http.Request) {
	links := make(map[string]string)
	links["/teams"] = "POINTS TABLE OF ALL THE TEAMS"
	links["/matches"] = "LIST OF ALL THE MATCHES"
	links["/matches/today"] = "LIST OF TODAY'S MATCHES"
	page := HomePage{Title: "FIFA WORLDCUP 2K18", Links: links}
	t, _ := template.ParseFiles("home.html")
	t.Execute(w, page)
}

func main() {
	server := http.Server{
		Addr: ":" + os.Getenv("PORT"), //
	}
	http.HandleFunc("/", fifa)
	http.HandleFunc("/teams", allteams)
	http.HandleFunc("/matches", fifaMatches)
	http.HandleFunc("/matches/today", todayMatches)
	//http.ListenAndServe("localhost:8000", nil)
	server.ListenAndServe()
}
