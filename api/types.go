package api

import "context"

type CTFdAPI interface {
	Login(ctx context.Context, user, password string) error
	GetChallenges(ctx context.Context) ([]ListChallenge, error)
	GetChallenge(ctx context.Context, id uint16) (*Challenge, error)
	SubmitFlag(ctx context.Context, id int, flag string) (*AttemptResult, error)
	GetScoreboard(ctx context.Context) ([]ScoreboardEntry, error)
}

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
