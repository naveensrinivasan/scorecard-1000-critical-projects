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
	f := openFile("all.csv")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	results := []Scorecard{}
	var syncMutex sync.Mutex
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
		//convert the score to float
		scoreFloat, err := strconv.ParseFloat(score, 64)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go func(dep string, score float64, criticality int) {
			defer wg.Done()
			scorecard, err := GetScore(dep)
			atomic.AddInt64(&ops, 1)
			if err != nil {
				return
			}
			syncMutex.Lock()
			defer syncMutex.Unlock()
			scorecard.Repo.CriticalityScore = score
			scorecard.Repo.Criticality = criticality
			results = append(results, scorecard)
		}(x, scoreFloat, i)
	}
	wg.Wait()
	//serialize the results to a file
	b, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("results.json", b, 0644)
	if err != nil {
		panic(err)
	}
}

func decodeJson(m map[string]interface{}) []string {
	values := make([]string, 0, len(m))
	for _, v := range m {
		switch vv := v.(type) {
		case map[string]interface{}:
			for _, value := range decodeJson(vv) {
				values = append(values, value)
			}
		case string:
			values = append(values, vv)
		case float64:
			values = append(values, strconv.FormatFloat(vv, 'f', -1, 64))
		case []interface{}:
			// Arrays aren't currently handled, since you haven't indicated that we should
			// and it's non-trivial to do so.
		case bool:
			values = append(values, strconv.FormatBool(vv))
		case nil:
			values = append(values, "nil")
		}
	}
	return values
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
