package unter

import (
	"fmt"
	"time"
)

type Kind uint

const (
	Shared Kind = iota + 1
	Private
)

func (k Kind) String() string {
	switch k {
	case Shared:
		return "shared"
	case Private:
		return "private"
	}

	return fmt.Sprintf("<Kind %d>", k)
}

type Ride struct {
	ID       string
	Driver   string
	Kind     Kind
	Start    time.Time
	End      time.Time
	Distance float64
}
