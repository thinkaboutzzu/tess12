package model

type LiveIDNResponses struct {
	Data struct {
		GetLiveStream []struct {
			Slug        string `json:"slug,omitempty"`
			Title       string `json:"title,omitempty"`
			ImageUrl    string `json:"image_url,omitempty"`
			ViewCount   int    `json:"view_count,omitempty"`
			LiveAt      string `json:"live_at,omitempty"`
			PlaybackUrl string `json:"playback_url,omitempty"`
			Status      string `json:"status"`
			Creator     struct {
				Name     string `json:"name,omitempty"`
				Username string `json:"username,omitempty"`
			} `json:"creator"`
		} `json:"getLivestreams,omitempty"`
	} `json:"data"`
}
