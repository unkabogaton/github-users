package models

type UpdateUserRequest struct {
	Login        string `json:"Login"`
	NodeID       string `json:"NodeID"`
	AvatarURL    string `json:"AvatarURL"`
	URL          string `json:"URL"`
	HTMLURL      string `json:"HTMLURL"`
	Type         string `json:"Type"`
	UserViewType string `json:"UserViewType"`
	SiteAdmin    bool   `json:"SiteAdmin"`
}
