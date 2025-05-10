package models

// Message types
const (
	// Basic connection messages
	MsgTypeLoginRequest      = "LOGIN_REQUEST"
	MsgTypeLoginResponse     = "LOGIN_RESPONSE"
	MsgTypeErrorNotification = "ERROR_NOTIFICATION"
)

// GenericMessage is the wrapper for all network messages
type GenericMessage struct {
	Type    string      `json:"type"`    // Message type
	Payload interface{} `json:"payload"` // Message payload
}

// LoginRequestPayload is the payload for a login request
type LoginRequestPayload struct {
	Username string `json:"username"` // Username for login
}

// LoginResponsePayload is the payload for a login response
type LoginResponsePayload struct {
	Success  bool   `json:"success"`  // Whether login was successful
	Message  string `json:"message"`  // Success or error message
	PlayerID string `json:"playerId"` // Optional player ID
}

// ErrorNotificationPayload is the payload for an error notification
type ErrorNotificationPayload struct {
	ErrorMessage string `json:"errorMessage"` // Error message
}
