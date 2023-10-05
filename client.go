package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	UserID string
	Token  string
}

func NewClient(userID, token string) *Client {
	return &Client{
		UserID: userID,
		Token:  token,
	}
}

func (c *Client) GetUser() (*User, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	user := &User{}
	err = json.Unmarshal(b, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *Client) CreateRepo(org, name, description string, public bool) error {
	url := "https://api.github.com/user/repos"
	if org != c.UserID {
		url = "https://api.github.com/orgs/" + org + "/repos"
	}
	b, err := json.Marshal(CreateRepoRequest{
		Name:        name,
		Description: description,
		Private:     !public,
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Content-Type", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d: body:\n%s", resp.StatusCode, string(body))
	}
	return nil
}

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

func (c *Client) DeleteRepo(org, name string) error {
	url := "https://api.github.com/repos/" + org + "/" + name
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "token "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status code: %d: body:\n%s", resp.StatusCode, string(body))
	}
	return nil
}
