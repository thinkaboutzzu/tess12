package model

type PMMessage struct {
	Id             string `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	Message        string `json:"message,omitempty"`
	ChannelId      string `json:"channelId,omitempty"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
	UserMessagesId string `json:"userMessagesId,omitempty"`
	Author         struct {
		GivenName    string `json:"givenName,omitempty"`
		FamilyName   string `json:"familyName,omitempty"`
		Nickname     string `json:"nickname,omitempty"`
		ProfileImage string `json:"profileImage,omitempty"`
	} `json:"author"`
}

type PMMessageResponses struct {
	Data struct {
		MessagesByUpdateAt struct {
			Items []PMMessage `json:"items,omitempty"`
		} `json:"messagesByUpdateAt,omitempty"`
	} `json:"data"`
}

type PMBirthdayMessage struct {
	Id      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Author  struct {
		Id           string `json:"id"`
		GivenName    string `json:"givenName,omitempty"`
		FamilyName   string `json:"familyName,omitempty"`
		Nickname     string `json:"nickname,omitempty"`
		ProfileImage string `json:"profileImage"`
	} `json:"author"`
}

type PMBirthdayResponses struct {
	Data struct {
		ListBirthdayMessageTemplates struct {
			Items []PMBirthdayMessage `json:"items,omitempty"`
		} `json:"listBirthdayMessageTemplates"`
	} `json:"data"`
}

type PMRanking struct {
	Name         string `json:"name,omitempty"`
	Count        int    `json:"count,omitempty"`
	TextCount    int    `json:"textCount"`
	ImageCount   int    `json:"imageCount"`
	VoiceCount   int    `json:"voiceCount"`
	Points       int    `json:"points"`
	ProfileImage string `json:"profileImage,omitempty"`
}

type PMMessageByUserId struct {
	Message string `json:"message,omitempty"`
}

type PMMessageByUserIdResponses struct {
	Data struct {
		GetUser struct {
			Messages struct {
				Items []PMMessageByUserId `json:"items,omitempty"`
			} `json:"messages"`
		} `json:"getUser"`
	} `json:"data"`
}
