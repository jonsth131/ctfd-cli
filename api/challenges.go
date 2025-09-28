package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *ApiClient) GetChallenges(ctx context.Context) ([]ListChallenge, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, challengesApiURL)
	resp, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var challenges ApiResponse[[]ListChallenge]
	if err := json.NewDecoder(resp.Body).Decode(&challenges); err != nil {
		return nil, err
	}

	if challenges.Success != true {
		return nil, ErrFailedFetchingChals
	}

	return challenges.Data, nil
}

func (c *ApiClient) GetChallenge(ctx context.Context, id uint16) (*Challenge, error) {
	u := fmt.Sprintf("%s%s/%d", c.baseUrl, challengesApiURL, id)

	resp, err := c.get(ctx, u)

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

func (c *ApiClient) SubmitFlag(ctx context.Context, id int, attempt string) (*AttemptResult, error) {
	resp, err := c.get(ctx, fmt.Sprintf("%s%s", c.baseUrl, challengesURL))

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

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewBuffer(jsonData))
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
