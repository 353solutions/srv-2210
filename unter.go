package unter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Kind uint8

const (
	Shared Kind = iota + 1
	Private
	maxKind
)

// String implement fmt.Stringer
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

var zeroTime time.Time

func (r Ride) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("missing ID")
	}

	if r.Driver == "" {
		return fmt.Errorf("missing driver")
	}

	if r.Kind <= 0 || r.Kind >= maxKind {
		return fmt.Errorf("bad kind: %d", r.Kind)
	}

	if r.Start.Equal(zeroTime) {
		return fmt.Errorf("missing start time")
	}

	if !r.End.Equal(zeroTime) && r.End.Before(r.Start) {
		return fmt.Errorf("end before start time (%v >= %v)", r.End, r.Start)
	}

	if r.Distance < 0 {
		return fmt.Errorf("negative distance: %f", r.Distance)
	}

	return nil
}

func NewID() string {
	return uuid.NewString()
}
