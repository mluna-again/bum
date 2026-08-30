package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ServerError = errors.New("server said no")

func clientRequestNew(body Pane) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	payload := bytes.NewReader(data)

	resp, err := http.Post(fmt.Sprintf("http://localhost:%s/new", port), "application/json", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%w: %q", string(b))
	}

	return nil
}
