package freebox_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
	"github.com/xfoucron/freebox-exporter/internal/freebox/freeboxtest"
)

func newClient(t *testing.T, srv *freeboxtest.Server, token string) *freebox.Client {
	t.Helper()
	c, err := freebox.New(context.Background(), srv.URL, freeboxtest.AppID, token, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestGet(t *testing.T) {
	srv := freeboxtest.NewServer(t, map[string]string{
		"connection/": `{"state":"up","type":"ethernet","rate_down":1500,"bytes_up":42}`,
	})
	c := newClient(t, srv, freeboxtest.AppToken)

	got, err := freebox.Get[freebox.Connection](context.Background(), c, "connection/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.State != "up" || got.RateDown != 1500 || got.BytesUp != 42 {
		t.Errorf("unexpected connection: %+v", got)
	}
}

func TestGetEmptyResult(t *testing.T) {
	srv := freeboxtest.NewServer(t, map[string]string{"storage/disk/": ""})
	c := newClient(t, srv, freeboxtest.AppToken)

	got, err := freebox.Get[[]freebox.Disk](context.Background(), c, "storage/disk/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d disks, want 0", len(got))
	}
}

func TestGetReusesSession(t *testing.T) {
	srv := freeboxtest.NewServer(t, map[string]string{"connection/": `{}`})
	c := newClient(t, srv, freeboxtest.AppToken)

	for range 3 {
		if _, err := freebox.Get[freebox.Connection](context.Background(), c, "connection/"); err != nil {
			t.Fatalf("Get: %v", err)
		}
	}
	if n := srv.Logins(); n != 1 {
		t.Errorf("logins = %d, want 1", n)
	}
}

func TestGetRenewsExpiredSession(t *testing.T) {
	srv := freeboxtest.NewServer(t, map[string]string{"connection/": `{"state":"up"}`})
	c := newClient(t, srv, freeboxtest.AppToken)

	if _, err := freebox.Get[freebox.Connection](context.Background(), c, "connection/"); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	srv.RevokeSessions()
	got, err := freebox.Get[freebox.Connection](context.Background(), c, "connection/")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if got.State != "up" {
		t.Errorf("state = %q, want up", got.State)
	}
	if n := srv.Logins(); n != 2 {
		t.Errorf("logins = %d, want 2", n)
	}
}

func TestGetBadToken(t *testing.T) {
	srv := freeboxtest.NewServer(t, map[string]string{"connection/": `{}`})
	c := newClient(t, srv, "wrong-token")

	_, err := freebox.Get[freebox.Connection](context.Background(), c, "connection/")
	var apiErr *freebox.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "invalid_token" {
		t.Errorf("err = %v, want invalid_token APIError", err)
	}
}

func TestGetUnknownResource(t *testing.T) {
	srv := freeboxtest.NewServer(t, nil)
	c := newClient(t, srv, freeboxtest.AppToken)

	_, err := freebox.Get[freebox.Connection](context.Background(), c, "nope/")
	var apiErr *freebox.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "no_such_resource" {
		t.Errorf("err = %v, want no_such_resource APIError", err)
	}
}
