package api

import "errors"

var (
	ErrInvalidUsername     = errors.New("username cannot be empty")
	ErrInvalidPassword     = errors.New("password cannot be empty")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrCaptchaRequired     = errors.New("CAPTCHA is required. Try to login using a browser.")
	ErrFailedFetchingChals = errors.New("failed to fetch challenges")
	ErrFailedFetchingBoard = errors.New("failed to fetch scoreboard")
)
