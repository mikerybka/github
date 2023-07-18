package github

type User struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}
