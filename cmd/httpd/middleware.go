package main

import (
	"context"
	"errors"
	"fmt"
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
				badLogins.Add(1)
				log.Printf("ERROR: <%s> [SEC] %q bad auth from %s", rid, login, r.RemoteAddr)
				http.Error(w, fmt.Sprintf("bad login (%s)", rid), http.StatusForbidden)
				return
			}
			log.Printf("INFO: <%s> [SEC] %q logged in from %s", rid, login, r.RemoteAddr)
			v.User = user
		} else {
			okLogins.Add(1)
			log.Printf("INFO: <%s> [SEC] no auth from %s", rid, r.RemoteAddr)
		}

		ctx := context.WithValue(r.Context(), ctxKey, &v)
		r = r.Clone(ctx)

		log.Printf("INFO: <%s> %s called", rid, r.URL.Path)
		start := time.Now()

		h.ServeHTTP(w, r)

		// after
		duration := time.Since(start)
		// exercise: Log the return HTTP status
		log.Printf("INFO: <%s> %s ended in %v", rid, r.URL.Path, duration)
	}

	return http.HandlerFunc(fn)
}
