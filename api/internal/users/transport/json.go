package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxRequestBodyBytes int64 = 64 << 10

func decodeJSONBody(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) error {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodyBytes,
	)

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}

		return fmt.Errorf("decode trailing request body: %w", err)
	}

	return nil
}
