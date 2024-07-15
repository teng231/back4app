package httpserver

type ErrorWrapResponse struct {
	// key error
	Error string `json:"error"`
	// Code error - can code of httpcode
	Code int `json:"code"`
	// Step description for 1 error step
	Step int `json:"step"`
	// full message - o
	Message string `json:"message"`
	// trace message - for debug mode
	Trace string `json:"trace"`
}
