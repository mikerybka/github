package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OAuthApp struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

func (a *OAuthApp) Login(code string) (*Client, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", a.ClientID, a.ClientSecret, code)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var oauthResponse struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
	err = json.NewDecoder(resp.Body).Decode(&oauthResponse)
	if err != nil {
		return nil, err
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
