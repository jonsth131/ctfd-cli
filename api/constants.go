package api

const (
	loginURL          = "/login"
	challengesApiURL  = "/api/v1/challenges"
	challengesURL     = "/challenges"
	flagAttemptApiURL = "/api/v1/challenges/attempt"
	scoreboardApiURL  = "/api/v1/scoreboard"

	cloudflareCAPTCHATitle = "Just a moment..."

	ctfdInvalidCredentials = "Your username or password is incorrect"

	sessionCookieName = "session"

	csrfTokenHeaderName = "Csrf-Token"

	errInvalidUsername          = "username cannot be empty"
	errInvalidPassword          = "password cannot be empty"
	errInvalidCredentials       = "invalid credentials"
	errFailedToGetLoginPage     = "failed to get login page"
	errFailedToCheckCAPTCHA     = "failed to check CAPTCHA"
	errCAPTCHARequired          = "CAPTCHA is required. Try to login using a browser."
	errFailedToExtractNonce     = "failed to extract nonce"
	errFailedToLogin            = "failed to login"
	errFailedToReadResponseBody = "failed to read response body"
	errEmptyResponseBody        = "empty response body"
	errFailedToExtractTitle     = "failed to extract title"
	errLoginCancelled           = "login cancelled"
	errLoginTimeout             = "login timed out"
	errNoSessionCookie          = "no session cookie found after login"
	errFailedFetchingChallenges = "failed to fetch challenges"
	errFailedFetchingChallenge  = "failed to fetch challenge"
	errFailedSubmittingFlag     = "failed to submit flag for challenge"
	errFailedFetchingScoreboard = "failed to fetch scoreboard"
)
