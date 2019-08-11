package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type player struct {
	ID   string `json:"id"`
	Age  string `json:"Age"`
	Name string `json:"name"`
}

type team struct {
	ID      uint64   `json:"id"`
	Name    string   `json:"name"`
	Players []player `json:"players"`
}

type data struct {
	Team team `json:"team"`
}

type teamDetails struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    data   `json:"data"`
}

type playerDetails struct {
	ID    string   `json:"id"`
	Age   string   `json:"age"`
	Teams []string `json:"teams"`
}

func main() {
	requiredTeams := []string{"Germany", "England", "France", "Spain", "Manchester United", "Arsenal", "Chelsea", "Barcelona", "Real Madrid", "Bayern Munich"}
	players := make(map[string]playerDetails)
	var playerNames []string
	var count uint
	count = math.MaxUint32
	var i uint
	var j int
	for i = 1; i < count && j < len(requiredTeams); i++ {
		resp, err := http.Get("https://vintagemonster.onefootball.com/api/teams/en/" + fmt.Sprint(i) + ".json")
		if err != nil {
			//log.Println(err)
			continue
		} else if resp.StatusCode != 200 {
			//log.Println("Response is not valid: ", resp.StatusCode)
			continue
		}
		defer resp.Body.Close()
		var teamDetails teamDetails
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			//log.Println(err)
			continue
		}
		err = json.Unmarshal(body, &teamDetails)
		if err != nil {
			//log.Println(err)
			continue
		}
		if !contains(requiredTeams, teamDetails.Data.Team.Name) {
			continue
		}
		j++
		for _, teamPlayer := range teamDetails.Data.Team.Players {
			if val, ok := players[teamPlayer.Name]; ok && teamPlayer.ID == val.ID {
				val.Teams = append(val.Teams, teamDetails.Data.Team.Name)
				players[teamPlayer.Name] = val
			} else {
				var teamPlayerDetails playerDetails
				teamPlayerDetails.Age = teamPlayer.Age
				teamPlayerDetails.ID = teamPlayer.ID
				teamPlayerDetails.Teams = append(teamPlayerDetails.Teams, teamDetails.Data.Team.Name)
				players[teamPlayer.Name] = teamPlayerDetails
				playerNames = append(playerNames, teamPlayer.Name)
			}
		}
	}
	cl := collate.New(language.English, collate.Loose)
	cl.SortStrings(playerNames)
	for index, playerName := range playerNames {
		println(strconv.Itoa(index) + ". " + playerName + "; " + players[playerName].Age + "; " + strings.Join(players[playerName].Teams, ", "))
	}
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
