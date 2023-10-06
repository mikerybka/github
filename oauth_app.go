package github

import (
	"encoding/json"
	"fmt"
	"io"
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
