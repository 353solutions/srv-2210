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
	User      User
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

type Role uint8

const (
	Viewer Role = iota + 1
	Writer
	Admin
)

type User struct {
	Login string
	Role  Role
}

func LoginUser(login, passwd string) (User, error) {
	// FIXME: Use a real service
	switch login {
	case "Bond":
		if passwd == "007" {
			return User{login, Writer}, nil
		}
	case "Q":
		if passwd == "s3cr3t" {
			return User{login, Viewer}, nil
		}
	}

	return User{}, ErrBadLogin
}

func HasRole(u User, roles ...Role) bool {
	for _, r := range roles {
		if u.Role == r {
			return true
		}
	}
	return false
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
			user, err := LoginUser(login, passwd)
			if err != nil {
				http.Error(w, "bad login", http.StatusForbidden)
				return
			}
			v.User = user
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
