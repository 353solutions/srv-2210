package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type keyType int

var ctxKey keyType = 1

type Values struct {
	RequestID string
	Login     string
}

func RequestValues(ctx context.Context) *Values {
	v, ok := ctx.Value(ctxKey).(*Values)
	if !ok {
		return nil
	}

	return v
}

func RequestID(ctx context.Context) string {
	v := RequestValues(ctx)
	if v == nil {
		return "XXXX"
	}
	return v.RequestID
}

var ErrBadLogin = errors.New("bad login")

// FIXME: Use a real service
// func LoginUser(login, passwd string) *User
func LoginUser(login, passwd string) error {
	ok := false
	switch login {
	case "Bond":
		ok = passwd == "007"
	case "Q":
		ok = passwd == "s3cr3t"
	}

	if !ok {
		return ErrBadLogin
	}
	return nil
}

// middleware
func topMiddleware(log *log.Logger, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// before
		rid := uuid.NewString()
		v := Values{
			RequestID: rid,
		}

		login, passwd, ok := r.BasicAuth()
		if ok {
			if err := LoginUser(login, passwd); err != nil {
				http.Error(w, "bad login", http.StatusForbidden)
				return
			}
			v.Login = login
		}

		ctx := context.WithValue(r.Context(), ctxKey, &v)
		r = r.Clone(ctx)

		log.Printf("%s called (rid = %s)", r.URL.Path, rid)
		start := time.Now()

		h.ServeHTTP(w, r)

		// after
		duration := time.Since(start)
		// exercise: Log the return HTTP status
		log.Printf("%s ended in %v (rid = %s)", r.URL.Path, duration, rid)
	}

	return http.HandlerFunc(fn)
}
