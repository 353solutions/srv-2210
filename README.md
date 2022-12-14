# Practical Go Services
ArdanLabs ∴  2022 <br />

Miki Tebeka
<i class="far fa-envelope"></i> [miki@353solutions.com](mailto:miki@353solutions.com), <i class="fab fa-twitter"></i> [@tebeka](https://twitter.com/tebeka), <i class="fab fa-linkedin-in"></i> [mikitebeka](https://www.linkedin.com/in/mikitebeka/), <i class="fab fa-blogger-b"></i> [blog](https://www.ardanlabs.com/blog/)

#### Shameless Plugs

- [Go Essential Training](https://www.linkedin.com/learning/go-essential-training/) - LinkedIn Learning
    - [Rest of classes](https://www.linkedin.com/learning/instructors/miki-tebeka)
- [Go Brain Teasers](https://pragprog.com/titles/d-gobrain/go-brain-teasers/) book

---

## Day 4

### Agenda

- Security best practices
- OWASP top 10
- The security mindset


[Terminal Log](_class/day-4.log)

### Links

- [casbin](https://github.com/casbin/casbin) - Authorization framework
- [golang.org/x/crypto/acme/autocert](https://pkg.go.dev/golang.org/x/crypto/acme/autocert) - Certificates from Let's Encrypt
- [Guide to HTTP timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) Rate limiting
- [OWASP top 10](https://owasp.org/www-project-top-ten/)
- [gosec](https://github.com/securego/gosec) - Security oriented linter
- [html/template](https://pkg.go.dev/html/template)
- [gorilla/handlers](https://pkg.go.dev/github.com/gorilla/handlers) - CORS

### Data & Other


- `curl -d@./_class/start.json http://localhost:8080/rides`
- `curl -uBond:007 -d@./_class/start.json http://localhost:8080/rides`

---

## Day 3

### Agenda

- Testing overview
- Testing handlers
- Running services in testing
- Mocking
- Performance optimization

[Terminal Log](_class/day-3.log)

### Links

- Linters
    - [staticcheck](https://staticcheck.io/)
    - [golangci-lint](https://golangci-lint.run/)
    - Write your own with [x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- Testing
    - [testing](https://pkg.go.dev/testing/)
    - [testify](https://pkg.go.dev/github.com/stretchr/testify) - Many test utilities (including suites & mocking)
    - [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
        - [testing/quick](https://pkg.go.dev/testing/quick) - Initial fuzzing library
    - [test containers](https://golang.testcontainers.org/)
    - [net/http/httptest](https://pkg.go.dev/net/http/httptest)
    - [How SQLite is Tested](https://www.sqlite.org/testing.html)
- Performance Optimization
    - Miki's [Optimization](_class/optimize.md) guidelines
    - [Amdahl's Law](https://en.wikipedia.org/wiki/Amdahl%27s_law) - Limits of concurrency
    - [Computer Latency at Human Scale](https://twitter.com/jordancurve/status/1108475342468120576/photo/1)
    - [Rules of Optimization Club](https://wiki.c2.com/?RulesOfOptimizationClub)
    - [Profiling Go Programs](https://go.dev/blog/pprof) In the Go blog
    - [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html) by Dave Cheney
    - [hey](https://github.com/rakyll/hey)
        - `go install github.com/rakyll/hey@latest`
        - `export PATH=$(go env GOPATH)/bin:${PATH}`

### Data & Other

- `go get github.com/rakyll/hey@latest`

---
## Day 2

### Agenda

- Databases
- Working with `database/sql`
- Working with redis
- Logging
- Monitoring 

[Terminal Log](day-2.log)

### Links

- [copier](https://github.com/jinzhu/copier) - Copy struct to another
- [multierr](https://pkg.go.dev/go.uber.org/multierr) - Combine errors
- Databases
    - [database/sql](https://pkg.go.dev/database/sql)
    - [sqlx](https://jmoiron.github.io/sqlx/)
    - [lig/pg](https://pkg.go.dev/github.com/lib/pq)
    - [sqlc](https://sqlc.dev/) - Generate code
    - [gorm](https://gorm.io/) - ORM
    - [pgcli](https://www.pgcli.com/) - PostgreSQL command line client
    - [go-redis](https://redis.uptrace.dev/)
    - [bbolt](https://github.com/etcd-io/bbolt)
- Logging & Metrics
    - [log](https://pkg.go.dev/log)
    - [uber/zap](https://pkg.go.dev/go.uber.org/zap)
    - [expvar](https://pkg.go.dev/expvar)
        - [expvarmon](https://github.com/divan/expvarmon)
    - [Open Telemetry](https://opentelemetry.io/docs/instrumentation/go/getting-started/)
    - [Let's talk about logging](https://dave.cheney.net/2015/11/05/lets-talk-about-logging) by Dave Cheney
    - [Prometheus metric types](https://prometheus.io/docs/concepts/metric_types/)

### Data & Other

- `docker exec -it <id> psql -U postgres`
    - or `pgcli -p 5432 -U postgres -h localhost`
- `docker exec -it <id> redis-cli`


---

## Day 1

### Agenda

- REST APIs
- Handlers
- JSON
- Middleware 

### Code

Clone the repo:

```
$ git clone https://github.com/353solutions/srv-2210.git unter
```

[Terminal Log](_class/day-1.log)

### Links

- Configuration
    - [conf](https://pkg.go.dev/github.com/ardanlabs/conf/v3)
    - [viper](https://github.com/spf13/viper) & [cobra](https://github.com/spf13/cobra)
- ArdanLabs [service](https://github.com/ardanlabs/service) - More complex system
- Validation
    - [cue](https://cuelang.org/)
    - [validator](https://pkg.go.dev/github.com/go-playground/validator/v10)
- REST APIs
    - [HTTP status cats](https://http.cat/)
    - [encoding/json](https://pkg.go.dev/encoding/json)
    - [net/http](https://pkg.go.dev/net/http)
    - [gorilla/mux](https://github.com/gorilla/mux) - HTTP router with more frills
    - [chi](https://github.com/go-chi/chi) - A nice web framework
        - Also Gin, Echo, Fiber, fasthttp, ...
- Setting up
    - [Go SDK](https://go.dev/dl/)
        - `go version`
    - And IDE such as [VSCode](https://code.visualstudio.com/) or [Goland](https://www.jetbrains.com/go/)
    - [git](https://git-scm.com/)
        - `git --version`
    - [Docker](https://www.docker.com/)
        - `docker pull postgres:15-alpine`
        - `docker pull redis:7-alpine`

### Design

```
[library]
layers:
    - business
    - foundations

front ends: [web] [cli] [gui]
```

Rules of thumb: imports go downward in the layers

Data Types: one per layer
    API
    Business
    Database


### Data & Other

- `curl -d'{"driver": "gomez", "kind": "private"}' http://localhost:8080/rides`
- `curl -d'{"distance": 3.14}' http://localhost:8080/rides/<id>/end`
