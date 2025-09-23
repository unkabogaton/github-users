package models

import "time"

type GitHubUser struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type User struct {
	Login        string    `db:"login"`
	ID           int       `db:"id"`
	NodeID       string    `db:"node_id"`
	AvatarURL    string    `db:"avatar_url"`
	URL          string    `db:"url"`
	HTMLURL      string    `db:"html_url"`
	Type         string    `db:"type"`
	UserViewType string    `db:"user_view_type"`
	SiteAdmin    bool      `db:"site_admin"`
	UpdatedAt    time.Time `db:"updated_at"`
	CreatedAt    time.Time `db:"created_at"`
}
