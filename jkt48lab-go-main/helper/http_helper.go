package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Fetch(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)
	resp, err := client.Do(req)
	return resp, err
}

func GraphQLIDN(page int) (*http.Response, error) {
	query, err := json.Marshal(map[string]any{
		"query": `
			query GetLivestream($category: String, $page: Int){
				getLivestreams(category: $category, page: $page){
					slug
					title
					image_url
					view_count
					playback_url
					room_identifier
					status
					scheduled_at
					live_at
					category {
						name
						slug
					}
					creator{
						name
						username
						uuid
					}
				}
			}
    	`,
		"variables": map[string]int{
			"page": page,
		},
	})
	gReq, _ := http.NewRequest("POST", "https://api.idn.app/graphql", bytes.NewBuffer(query))
	gReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(gReq)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func GraphQLRequest(url string, query []byte, authorization string) (*http.Response, error) {
	gReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(query))
	gReq.Header.Set("Authorization", authorization)
	gReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(gReq)
	if err != nil {
		return nil, err
	}
	return resp, err
}
