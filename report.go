package unter

import "time"

const (
	minFee  = 250 // ¢
	perMile = 250
	perHour = 3000
)

// RideFee returns the ride fee in ¢
func RideFee(duration time.Duration, distance float64, shared bool) int {
	m := perMile * distance
	h := perHour * float64(duration/time.Minute/60)

	fee := max(m, h)
	fee = max(fee, minFee)
	if shared {
		fee = 0.9 * float64(fee)
	}

	return int(fee)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

type Report struct {
	Driver   string
	NumRides int
	Payment  int
}

func ByDriver(rides []Ride) []Report {
	rs := make(map[string]*Report) // driver -> report
	for _, r := range rides {
		rp, ok := rs[r.Driver]
		if !ok {
			rp = &Report{
				Driver: r.Driver,
			}
			rs[r.Driver] = rp
		}
		rp.NumRides++

		duration := r.End.Sub(r.Start)
		rp.Payment += RideFee(duration, r.Distance, r.Kind == Shared) - 30
	}

	reports := make([]Report, 0, len(rs))
	for _, rp := range rs {
		reports = append(reports, *rp)
	}
	return reports
}
