package unter

import "time"

type Report struct {
	Driver   string
	NumRides int
	Payment  int
}

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

func findReport(driver string, rs []Report) int {
	for i, r := range rs {
		if r.Driver == driver {
			return i
		}
	}

	return -1
}

func ByDriver(rides []Ride) []Report {
	var rs []Report
	for _, r := range rides {
		i := findReport(r.Driver, rs)
		if i == -1 {
			rs = append(rs, Report{Driver: r.Driver})
			i = len(rs) - 1
		}
		rs[i].NumRides++

		duration := r.End.Sub(r.Start)
		rs[i].Payment += RideFee(duration, r.Distance, r.Kind == Shared) - 30
	}

	return rs
}
