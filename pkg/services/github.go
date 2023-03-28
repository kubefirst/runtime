package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/kubefirst/runtime/pkg/helpers"
	"github.com/kubefirst/runtime/pkg/httpCommon"
)

type GitHubService struct {
	httpClient httpCommon.HTTPDoer
}

// gitHubAccessCode host OAuth data
type gitHubAccessCode struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// NewGitHubService instantiate a new GitHub service
func NewGitHubService(httpClient httpCommon.HTTPDoer) *GitHubService {
	return &GitHubService{
		httpClient: httpClient,
	}
}

// CheckUserCodeConfirmation checks if the user gave permission to the device flow request
func (service GitHubService) CheckUserCodeConfirmation(deviceCode string) (string, error) {

	gitHubAccessTokenURL := "https://github.com/login/oauth/access_token"

	jsonData, err := json.Marshal(map[string]string{
		"client_id":   helpers.GitHubOAuthClientId,
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, gitHubAccessTokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil
	}

	req.Header.Add("Content-Type", helpers.JSONContentType)
	req.Header.Add("Accept", helpers.JSONContentType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("waiting user to authorize at GitHub page..., current status code = %d", res.StatusCode)
		return "", errors.New("unable to issue a GitHub token")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}

	var gitHubAccessToken gitHubAccessCode
	err = json.Unmarshal(body, &gitHubAccessToken)
	if err != nil {
		log.Println(err)
	}

	return gitHubAccessToken.AccessToken, nil
}