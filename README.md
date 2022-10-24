# Practical Go Services
ArdanLabs ∴  2022 <br />

Miki Tebeka
<i class="far fa-envelope"></i> [miki@353solutions.com](mailto:miki@353solutions.com), <i class="fab fa-twitter"></i> [@tebeka](https://twitter.com/tebeka), <i class="fab fa-linkedin-in"></i> [mikitebeka](https://www.linkedin.com/in/mikitebeka/), <i class="fab fa-blogger-b"></i> [blog](https://www.ardanlabs.com/blog/)

#### Shameless Plugs

- [Go Essential Training](https://www.linkedin.com/learning/go-essential-training/) - LinkedIn Learning
    - [Rest of classes](https://www.linkedin.com/learning/instructors/miki-tebeka)
- [Go Brain Teasers](https://pragprog.com/titles/d-gobrain/go-brain-teasers/) book



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

- REST APIs
    - [HTTP status cats](https://http.cat/)
    - [encoding/json](https://pkg.go.dev/encoding/json)
    - [net/http](https://pkg.go.dev/net/http)
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
