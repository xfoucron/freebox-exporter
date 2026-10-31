package freebox

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // HMAC-SHA1 is mandated by the Freebox login protocol
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrAuthorizationDenied  = errors.New("authorization denied on the freebox")
	ErrAuthorizationTimeout = errors.New("authorization timed out on the freebox")
)

func (c *Client) login(ctx context.Context) (string, error) {
	var challenge struct {
		Challenge string `json:"challenge"`
	}
	if err := c.roundTrip(ctx, http.MethodGet, c.baseURL+"login/", "", nil, &challenge, true); err != nil {
		return "", err
	}

	mac := hmac.New(sha1.New, []byte(c.token))
	mac.Write([]byte(challenge.Challenge))

	request := map[string]string{
		"app_id":   c.appID,
		"password": hex.EncodeToString(mac.Sum(nil)),
	}
	var session struct {
		SessionToken string `json:"session_token"`
	}
	if err := c.roundTrip(ctx, http.MethodPost, c.baseURL+"login/session/", "", request, &session, true); err != nil {
		return "", err
	}
	return session.SessionToken, nil
}

type Application struct {
	ID      string `json:"app_id"`
	Name    string `json:"app_name"`
	Version string `json:"app_version"`
	Device  string `json:"device_name"`
}

func (c *Client) Authorize(ctx context.Context, app Application) (token string, trackID int, err error) {
	var result struct {
		AppToken string `json:"app_token"`
		TrackID  int    `json:"track_id"`
	}
	if err := c.roundTrip(ctx, http.MethodPost, c.baseURL+"login/authorize/", "", app, &result, true); err != nil {
		return "", 0, err
	}
	return result.AppToken, result.TrackID, nil
}

func (c *Client) WaitGranted(ctx context.Context, trackID int, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	path := c.baseURL + "login/authorize/" + strconv.Itoa(trackID)

	for {
		var result struct {
			Status string `json:"status"`
		}
		if err := c.roundTrip(ctx, http.MethodGet, path, "", nil, &result, true); err != nil {
			return err
		}
		switch result.Status {
		case "granted":
			return nil
		case "pending":
		case "denied":
			return ErrAuthorizationDenied
		case "timeout":
			return ErrAuthorizationTimeout
		default:
			return fmt.Errorf("unexpected authorization status %q", result.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
