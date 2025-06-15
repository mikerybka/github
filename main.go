package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mikerybka/util"
)

func main() {
	token := util.RequireEnvVar("GITHUB_TOKEN")
	http.HandleFunc("POST /api/webhooks", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(os.Stdout, r.Body)
	})

	http.HandleFunc("POST /api/create-repo", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}

		// Decode the JSON body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create the repo via GitHub API
		apiURL := "https://api.github.com/user/repos"
		payload := fmt.Sprintf(`{"name":"%s"}`, req.Name)
		reqBody := strings.NewReader(payload)

		reqGitHub, err := http.NewRequest("POST", apiURL, reqBody)
		if err != nil {
			http.Error(w, "Failed to create GitHub request", http.StatusInternalServerError)
			return
		}
		reqGitHub.Header.Set("Authorization", "token "+token)
		reqGitHub.Header.Set("Accept", "application/vnd.github+json")
		reqGitHub.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(reqGitHub)
		if err != nil {
			http.Error(w, "GitHub API request failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	})
	http.HandleFunc("POST /api/write-files", func(w http.ResponseWriter, r *http.Request) {

	})
}

type OAuthApp struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

func (a *OAuthApp) Login(code string) (*Client, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", a.ClientID, a.ClientSecret, code)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %s", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %s", err)
	}
	defer resp.Body.Close()
	var oauthResponse struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %s", err)
	}
	fmt.Println(string(b))
	err = json.Unmarshal(b, &oauthResponse)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response: %s", err)
	}
	token := oauthResponse.AccessToken
	return &Client{
		Token: token,
	}, nil
}

type UserResponse struct {
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

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
		return nil, fmt.Errorf("creating request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %s", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %s", err)
	}
	fmt.Println(string(b))
	user := &User{}
	err = json.Unmarshal(b, user)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response: %s", err)
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

type User struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type WebhookHandler func(w *Webhook) error

func (h WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wh Webhook
	json.NewDecoder(r.Body).Decode(&wh)
	err := h(&wh)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
