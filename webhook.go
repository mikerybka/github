package github

type Webhook struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
}

type Repository struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
}
