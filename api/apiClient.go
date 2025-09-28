package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type ApiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type Challenge struct {
	Id             uint32   `json:"id"`
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Value          uint32   `json:"value"`
	Solves         uint32   `json:"solves"`
	SolvedByMe     bool     `json:"solved_by_me"`
	Category       string   `json:"category"`
	Files          []string `json:"files"`
	ConnectionInfo string   `json:"connection_info"`
	Tags           []string `json:"tags"`
	Attempts       int      `json:"attempts"`
	MaxAttempts    int      `json:"max_attempts"`
	Hints          []struct {
		Id   int `json:"id"`
		Cost int `json:"cost"`
	} `json:"hints"`
}

type ListChallenge struct {
	Id         uint32 `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Value      uint32 `json:"value"`
	Solves     uint32 `json:"solves"`
	SolvedByMe bool   `json:"solved_by_me"`
	Category   string `json:"category"`
}

type ScoreboardEntry struct {
	Position    uint32 `json:"pos"`
	AccountID   uint32 `json:"account_id"`
	AccountURL  string `json:"account_url"`
	AccountType string `json:"account_type"`
	OAuthID     uint32 `json:"oauth_id"`
	Name        string `json:"name"`
	Score       int32  `json:"score"`
	BracketID   uint32 `json:"bracket_id"`
	BracketName string `json:"bracket_name"`
	Members     []struct {
		ID          uint32 `json:"id"`
		OAuthID     uint32 `json:"oauth_id"`
		Name        string `json:"name"`
		Score       int32  `json:"score"`
		BracketID   uint32 `json:"bracket_id"`
		BracketName string `json:"bracket_name"`
	} `json:"members"`
}

type AttemptResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AttemptRequest struct {
	ChallengeId int    `json:"challenge_id"`
	Submission  string `json:"submission"`
}

type ApiClient struct {
	client  *http.Client
	baseUrl *url.URL
}

func NewApiClient(u string) (*ApiClient, error) {
	ur, err := parseBaseUrl(u)

	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
	}

	return &ApiClient{
		client:  client,
		baseUrl: ur,
	}, nil
}

func (api *ApiClient) SetCookie(cookie *http.Cookie) {
	api.client.Jar.SetCookies(api.baseUrl, []*http.Cookie{cookie})
}

func (c *ApiClient) Login(name string, password string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf(errInvalidUsername)
	}
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf(errInvalidPassword)
	}

	bodyString, err := c.getLoginPageBody()
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToGetLoginPage, err)
	}

	res, err := checkCAPTCHA(bodyString)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToCheckCAPTCHA, err)
	}

	if res {
		return fmt.Errorf(errCAPTCHARequired)
	}

	nonce, err := extractNonce(bodyString)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToExtractNonce, err)
	}

	err = c.performLogin(name, password, nonce)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToLogin, err)
	}

	return nil
}

func (c *ApiClient) getLoginPageBody() (string, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, loginURL)

	resp, err := c.client.Get(u)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %v", errFailedToReadResponseBody, err)
	}

	bodyString := string(bodyBytes)

	if bodyString == "" {
		return "", fmt.Errorf(errEmptyResponseBody)
	}

	return bodyString, nil
}

func checkCAPTCHA(body string) (bool, error) {
	title, err := extractTitle(body)
	if err != nil {
		return false, fmt.Errorf("%s: %w", errFailedToExtractTitle, err)
	}

	if strings.Contains(title, cloudflareCAPTCHATitle) {
		return true, nil
	}

	return false, nil
}

func (c *ApiClient) performLogin(username string, password string, nonce string) error {
	body := url.Values{
		"username": {username},
		"password": {password},
		"nonce":    {nonce},
	}

	resp, err := c.client.PostForm(fmt.Sprintf("%s%s", c.baseUrl, loginURL), body)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf(errLoginCancelled)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf(errLoginTimeout)
		}
		return fmt.Errorf("%s: %v", errFailedToLogin, err)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %v", errFailedToReadResponseBody, err)
	}
	bodyString := string(bodyBytes)

	if strings.Contains(bodyString, ctfdInvalidCredentials) {
		return fmt.Errorf(errInvalidCredentials)
	}

	cookies := resp.Request.CookiesNamed(sessionCookieName)
	if len(cookies) == 0 {
		return fmt.Errorf(errNoSessionCookie)
	}

	sessionCookie := cookies[0]
	c.SetCookie(sessionCookie)

	return nil
}

func (c *ApiClient) GetChallenges() ([]ListChallenge, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, challengesApiURL)

	resp, err := c.client.Get(u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var challenges ApiResponse[[]ListChallenge]
	if err := json.NewDecoder(resp.Body).Decode(&challenges); err != nil {
		return nil, err
	}

	if challenges.Success != true {
		return nil, fmt.Errorf(errFailedFetchingChallenges)
	}

	return challenges.Data, nil
}

func (c *ApiClient) GetChallenge(id uint16) (*Challenge, error) {
	u := fmt.Sprintf("%s%s/%d", c.baseUrl, challengesApiURL, id)

	resp, err := c.client.Get(u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var challenge ApiResponse[Challenge]
	if err := json.NewDecoder(resp.Body).Decode(&challenge); err != nil {
		return nil, err
	}

	if challenge.Success != true {
		return nil, fmt.Errorf("%s: %d", errFailedFetchingChallenge, id)
	}

	return &challenge.Data, nil
}

func (c *ApiClient) SubmitFlag(id int, attempt string) (*AttemptResult, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s%s", c.baseUrl, challengesURL))

	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errFailedToReadResponseBody, err)
	}
	bodyString := string(bodyBytes)

	nonce, err := extractCSRFToken(bodyString)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s%s", c.baseUrl, flagAttemptApiURL)

	request := AttemptRequest{
		ChallengeId: id,
		Submission:  attempt,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfTokenHeaderName, nonce)

	resp, err = c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response ApiResponse[AttemptResult]
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Success != true {
		return nil, fmt.Errorf("%s: %d", errFailedSubmittingFlag, id)
	}

	return &response.Data, nil
}

func (c *ApiClient) GetScoreboard() ([]ScoreboardEntry, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, scoreboardApiURL)

	resp, err := c.client.Get(u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var scoreboard ApiResponse[[]ScoreboardEntry]
	if err := json.NewDecoder(resp.Body).Decode(&scoreboard); err != nil {
		return nil, err
	}

	if scoreboard.Success != true {
		return nil, fmt.Errorf(errFailedFetchingScoreboard)
	}

	return scoreboard.Data, nil
}
