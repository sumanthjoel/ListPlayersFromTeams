package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Player contains detailed information of a player from a team
// We only read information that we need
type Player struct {
	ID   string `json:"id"`
	Age  string `json:"Age"`
	Name string `json:"name"`
}

// TeamDetails contains the detailed information about each team.
// We only read information that we require.
type TeamDetails struct {
	ID      uint     `json:"id"`
	Name    string   `json:"name"`
	Players []Player `json:"players"`
}

// Data is the required information from APIResponse
type Data struct {
	TeamDetails TeamDetails `json:"team"`
}

// APIResponse is the content read from the API endpoint
type APIResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

// playerDetails is the value for the map of required players' information
// players = make(map[string]playerDetails)
type playerDetails struct {
	ID    string   `json:"id"`
	Age   string   `json:"age"`
	Teams []string `json:"teams"`
}

// Mutex to safely read the map concurrently
var (
	players map[string]playerDetails
	lock    sync.RWMutex
)

func main() {
	var wg sync.WaitGroup
	requiredTeams := []string{"Germany", "England", "France", "Spain", "Manchester United", "Arsenal", "Chelsea", "Barcelona", "Real Madrid", "Bayern Munich"}
	players = make(map[string]playerDetails)
	var playerNames []string
	// identify the maximum value for uint
	const MaxUint = ^uint(0)
	var i uint
	var noOfTeamsFound int
	// When noOfTeamsFound matches len(requiredTeams), we got all we need
	for ; i < MaxUint && noOfTeamsFound < len(requiredTeams); i++ {
		resp, err := http.Get("https://vintagemonster.onefootball.com/api/teams/en/" + fmt.Sprint(i) + ".json")
		if err != nil {
			continue
		} else if resp.StatusCode != 200 {
			continue
		}
		defer resp.Body.Close()
		var apiResponse APIResponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			continue
		}
		if !contains(requiredTeams, apiResponse.Data.TeamDetails.Name) {
			continue
		}
		noOfTeamsFound++
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, teamPlayer := range apiResponse.Data.TeamDetails.Players {
				lock.RLock()
				if val, ok := players[teamPlayer.Name]; ok && teamPlayer.ID == val.ID {
					lock.RUnlock()
					val.Teams = append(val.Teams, apiResponse.Data.TeamDetails.Name)
					lock.Lock()
					players[teamPlayer.Name] = val
					lock.Unlock()
				} else {
					lock.RUnlock()
					var teamPlayerDetails playerDetails
					teamPlayerDetails.Age = teamPlayer.Age
					teamPlayerDetails.ID = teamPlayer.ID
					teamPlayerDetails.Teams = append(teamPlayerDetails.Teams, apiResponse.Data.TeamDetails.Name)
					lock.Lock()
					players[teamPlayer.Name] = teamPlayerDetails
					lock.Unlock()
					playerNames = append(playerNames, teamPlayer.Name)
				}
			}
		}()
	}
	wg.Wait()
	cl := collate.New(language.English, collate.Loose)
	cl.SortStrings(playerNames)
	lock.RLock()
	for index, playerName := range playerNames {
		println(strconv.Itoa(index+1) + ". " + playerName + "; " + players[playerName].Age + "; " + strings.Join(players[playerName].Teams, ", "))
	}
	lock.RUnlock()
}

// contains returns true if the string is in the array, else returns false
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
