package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	Username string
	Token    string
	Repos    []string
	Users    []string
}

type User struct {
	Login string
}

type PullRequest struct {
	Html_url string
	Title    string
	User     User
	Draft    bool
}

var Separator string = ","

func loadConfig(path string) *Config {
	config := &Config{}

	data, err := ioutil.ReadFile(path)
	checkError(err)

	err = json.Unmarshal(data, config)
	checkError(err)

	return config
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func draftFilter() func(pr PullRequest) bool {
	return func(pr PullRequest) bool {
		return pr.Draft
	}
}

func userFilter(users []string) func(pr PullRequest) bool {
	return func(pr PullRequest) bool {
		for _, user := range users {
			if pr.User.Login == user {
				return true
			}
		}
		return false
	}
}

func writeAllLines(filename string, lines []string) {
	file, err := os.Create(filename)
	checkError(err)

	for _, line := range lines {
		_, err := file.WriteString(line)
		checkError(err)
	}

	err = file.Close()
	checkError(err)
}

func main() {
	config := loadConfig("config.json")

	var openprLines []string
	var draftprLines []string
	for _, repo := range config.Repos {
		fmt.Printf("Started to process repo: %s\n", repo)

		query := make(url.Values)
		query.Add("per_page", "100")

		var requestUrl url.URL
		requestUrl.Scheme = "https"
		requestUrl.Host = "api.github.com"
		requestUrl.Path = "repos/networkservicemesh/" + repo + "/pulls"
		requestUrl.User = url.UserPassword(config.Username, config.Token)
		requestUrl.RawQuery = query.Encode()

		response, err := http.Get(requestUrl.String())
		checkError(err)

		responseData, err := ioutil.ReadAll(response.Body)
		checkError(err)

		var PRs []PullRequest
		err = json.Unmarshal(responseData, &PRs)
		checkError(err)

		inUsers := userFilter(config.Users)
		isDraft := draftFilter()

		for _, pr := range PRs {
			if inUsers(pr) {
				if !isDraft(pr) {
					openprLines = append(openprLines, pr.User.Login+Separator+pr.Title+Separator+pr.Html_url+"\n")
				} else {
					draftprLines = append(draftprLines, pr.User.Login+Separator+pr.Title+Separator+pr.Html_url+"\n")
				}
			}
		}
	}

	writeAllLines("open_prs.csv", openprLines)
	writeAllLines("draft_prs.csv", draftprLines)
}
