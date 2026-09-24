package httpresponse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	requestID string,
	details []FieldError,
) {
	response := ErrorResponse{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			RequestID: requestID,
			Details:   details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf(
			"error writing response: request_id=%s error=%v",
			requestID,
			err,
		)
	}
}

type DataResponse struct {
	Data any `json:"data"`
}

func WriteData(
	w http.ResponseWriter,
	status int,
	data any,
) error {
	reponse := DataResponse{Data: data}

	body, err := json.Marshal(reponse)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("error writing response: %v", err)
	}

	return nil
}
