package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Scorecard struct {
	Repo struct {
		Name             string  `json:"name"`
		CriticalityScore float64 `json:"criticalityScore"`
		Criticality      int     `json:"criticality"`
	} `json:"repo"`
	Score  float64 `json:"score"`
	Checks []struct {
		Name  string `json:"name"`
		Score int    `json:"score,omitempty"`
	} `json:"checks"`
}

func main() {
	f := openFile("1000_critical_projects.csv")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	var wg sync.WaitGroup
	var ops int64
	for i := 0; i < 1000; i++ {
		scanner.Scan()
		line := scanner.Text()
		//split the line by comma and take the first column
		x := strings.Split(line, ",")[0]
		x = strings.TrimLeft(x, "https://")
		// the last column in the csv file is empty is the score
		items := strings.Split(line, ",")
		score := items[len(items)-1]
		if score == "" {
			continue
		}
		//convert the score to float
		scoreFloat, err := strconv.ParseFloat(score, 64)
		if err != nil {
			log.Fatal(fmt.Errorf("error converting score to float: %w %d %s", err, i, score))
		}
		wg.Add(1)
		go func(dep string, score float64, criticality int) {
			defer wg.Done()
			scorecard, err := GetScore(dep)
			atomic.AddInt64(&ops, 1)
			if err != nil {
				return
			}
			scorecard.Repo.CriticalityScore = score
			scorecard.Repo.Criticality = criticality
			b, err := json.Marshal(scorecard)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(fmt.Sprintf("results/%d.json", criticality), b, 0644)
			if err != nil {
				return
			}
		}(x, scoreFloat, i+1)
	}
	wg.Wait()
}

func openFile(filename string) *os.File {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

// GetScore returns the scorecard score for a given repo.
func GetScore(repo string) (Scorecard, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.securityscorecards.dev/projects/%s", repo), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Scorecard{}, err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Scorecard{}, err
	}
	var scorecard Scorecard
	err = json.Unmarshal(result, &scorecard)
	if err != nil {
		return Scorecard{}, err
	}
	return scorecard, nil
}
