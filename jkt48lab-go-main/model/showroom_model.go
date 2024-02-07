package model

type LiveShowroomResponses struct {
	OnLives []struct {
		Lives []struct {
			RoomUrlKey       string `json:"room_url_key,omitempty"`
			StartedAt        int    `json:"started_at,omitempty"`
			RoomId           int    `json:"room_id,omitempty"`
			Image            string `json:"image,omitempty"`
			ViewNum          int    `json:"view_num,omitempty"`
			MainName         string `json:"main_name,omitempty"`
			PremiumRoomType  int    `json:"premium_room_type,omitempty"`
			StreamingUrlList []struct {
				Url string `json:"url"`
			} `json:"streaming_url_list"`
		} `json:"lives"`
	} `json:"onlives"`
}

type LiveShowroomStreamingUrlResponses struct {
	StreamingUrlList []struct {
		Url string `json:"url,omitempty"`
	} `json:"streaming_url_list,omitempty"`
}

type LiveShowroomGiftListResponses struct {
	Enquete []struct {
		GiftId   int    `json:"gift_id,omitempty"`
		GiftType int    `json:"gift_type,omitempty"`
		Image    string `json:"image,omitempty"`
		Free     bool   `json:"free,omitempty"`
		Point    int    `json:"point,omitempty"`
		GiftName string `json:"gift_name,omitempty"`
	} `json:"enquete,omitempty"`
	Normal []struct {
		GiftId   int    `json:"gift_id,omitempty"`
		GiftType int    `json:"gift_type,omitempty"`
		Image    string `json:"image,omitempty"`
		Free     bool   `json:"free,omitempty"`
		Point    int    `json:"point,omitempty"`
		GiftName string `json:"gift_name,omitempty"`
	} `json:"normal,omitempty"`
}

type LiveShowroomGiftLogResponses struct {
	GiftLog []struct {
		Name   string `json:"name,omitempty"`
		Aft    int    `json:"aft,omitempty"`
		Num    int    `json:"num,omitempty"`
		Image  string `json:"image,omitempty"`
		GiftId int    `json:"gift_id,omitempty"`
		UserId int    `json:"user_id,omitempty"`
	} `json:"gift_log,omitempty"`
}

type LiveShowroomGift struct {
	GiftType int
	Image    string
	Free     bool
	Point    int
	GiftName string
	Num      int
	UserId   int
}
