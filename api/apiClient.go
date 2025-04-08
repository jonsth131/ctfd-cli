package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type ApiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type Challenge struct {
	Id          uint32 `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       uint32 `json:"value"`
	Solves      uint32 `json:"solves"`
	SolvedByMe  bool   `json:"solved_by_me"`
	Category    string `json:"category"`
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
	ur, err := getBaseUrl(u)

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

func getBaseUrl(u string) (*url.URL, error) {
	ur, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	u = fmt.Sprintf("%s://%s", ur.Scheme, ur.Host)

	ur, err = url.Parse(u)
	if err != nil {
		return nil, err
	}

	return ur, nil
}

func (c *ApiClient) Login(name string, password string) (bool, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/login", c.baseUrl))

	if err != nil {
		return false, fmt.Errorf("Failed to get login page: %v", err)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("Failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)

	regex := regexp.MustCompile(`<input id="nonce" name="nonce" type="hidden" value="([a-f0-9]*)">`)

	nonce := regex.FindStringSubmatch(bodyString)[1]

	body := url.Values{
		"name":     {name},
		"password": {password},
		"nonce":    {nonce},
	}

	resp, err = c.client.PostForm(fmt.Sprintf("%s/login", c.baseUrl), body)

	if err != nil {
		return false, fmt.Errorf("Failed to login: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Failed to login: %v", resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("Failed to read response body: %v", err)
	}
	bodyString = string(bodyBytes)

	if strings.Contains(bodyString, "Your username or password is incorrect") {
		return false, fmt.Errorf("Invalid credentials")
	}

	sessionCookie := resp.Request.CookiesNamed("session")[0]

	c.SetCookie(sessionCookie)

	return true, nil
}

func (c *ApiClient) GetChallenges() ([]Challenge, error) {
	u := fmt.Sprintf("%s/api/v1/challenges", c.baseUrl)

	resp, err := c.client.Get(u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var challenges ApiResponse[[]Challenge]
	if err := json.NewDecoder(resp.Body).Decode(&challenges); err != nil {
		return nil, err
	}

	if challenges.Success != true {
		return nil, fmt.Errorf("Error fetching challenges")
	}

	return challenges.Data, nil
}

func (c *ApiClient) GetChallenge(id uint16) (*Challenge, error) {
	u := fmt.Sprintf("%s/api/v1/challenges/%d", c.baseUrl, id)

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
		return nil, fmt.Errorf("Error fetching challenge %d", id)
	}

	return &challenge.Data, nil
}

func (c *ApiClient) SubmitFlag(id int, attempt string) (*AttemptResult, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/challenges", c.baseUrl))

	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	regex := regexp.MustCompile(`'csrfNonce': "([a-f0-9]*)",`)

	nonce := regex.FindStringSubmatch(bodyString)[1]

	u := fmt.Sprintf("%s/api/v1/challenges/attempt", c.baseUrl)

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
	req.Header.Set("Csrf-Token", nonce)

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
		return nil, fmt.Errorf("Error submitting flag for challenge %d", id)
	}

	return &response.Data, nil
}

func (c *ApiClient) GetScoreboard() ([]ScoreboardEntry, error) {
	u := fmt.Sprintf("%s/api/v1/scoreboard", c.baseUrl)

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
		return nil, fmt.Errorf("Error fetching scoreboard")
	}

	return scoreboard.Data, nil
}
