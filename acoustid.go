package acoustid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type AcoustIDRequest struct {
	Fingerprint string `json:"fingerprint"`
	Duration    int    `json:"duration"`
	ApiKey      string `json:"client"`
	Metadata    string `json:"meta"`
}

type Result struct {
	ID string `json:"id"`

	Recordings []struct {
		Artists []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artists"`

		ReleaseGroups []struct {
			Type           string   `json:"type"`
			ID             string   `json:"id"`
			Title          string   `json:"title"`
			SecondaryTypes []string `json:"secondarytypes"`
		} `json:"releasegroups"`

		Duration float64 `json:"duration"`
		ID       string  `json:"id"`
		Title    string  `json:"title"`
	} `json:"recordings"`

	Score float64 `json:"score"`
}

type AcoustIDResponse struct {
	Results []Result `json:"results"`
	Status  string   `json:"status"`
}

func (a *AcoustIDRequest) Do() AcoustIDResponse {
	client := http.Client{}
	response, err := client.PostForm("http://api.acoustid.org/v2/lookup", a.PostValues())
	check(err)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	aidResp := AcoustIDResponse{}
	err = json.Unmarshal(body, &aidResp)
	check(err)

	return aidResp
}

func (a *AcoustIDRequest) PostValues() url.Values {
	query := fmt.Sprintf(
		"client=%s&duration=%d&meta=%s&fingerprint=%s",
		a.ApiKey,
		a.Duration,
		a.Metadata,
		a.Fingerprint)

	values, err := url.ParseQuery(query)
	check(err)
	return values
}
