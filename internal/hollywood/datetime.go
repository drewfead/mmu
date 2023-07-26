package hollywood

import "time"

const (
	monthAndDayFormat = "Monday, January 2"
	clockTimeFormat   = "3:04 PM"
)

func ParseDateTime(monthAndDay, clockTime string, timeZone *time.Location) (time.Time, bool) {
	d, err := time.ParseInLocation(monthAndDayFormat, monthAndDay, timeZone)
	if err != nil {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(clockTimeFormat, clockTime, timeZone)
	if err != nil {
		return d, false
	}
	offset := t.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, timeZone))
	return d.Add(offset), true
}
