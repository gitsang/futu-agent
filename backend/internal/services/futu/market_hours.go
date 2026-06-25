package futu

import (
	"fmt"
	"time"
)

type MarketHours struct {
	Open  int
	Close int
}

var marketSchedules = map[string][]MarketHours{
	"A股": {
		{Open: 930, Close: 1130},
		{Open: 1300, Close: 1500},
	},
	"港股": {
		{Open: 930, Close: 1200},
		{Open: 1300, Close: 1600},
	},
	"美股": {
		{Open: 930, Close: 1600},
	},
}

var marketTimezones = map[string]*time.Location{
	"CN": time.FixedZone("CST", 8*3600),
	"HK": time.FixedZone("HKT", 8*3600),
	"US": func() *time.Location {
		loc, _ := time.LoadLocation("America/New_York")
		if loc == nil {
			return time.FixedZone("EST", -5*3600)
		}
		return loc
	}(),
}

var marketNames = map[string]string{
	"CN": "A股",
	"HK": "港股",
	"US": "美股",
}

func IsMarketOpen(market string) bool {
	loc, ok := marketTimezones[market]
	if !ok {
		return false
	}

	now := time.Now().In(loc)
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	marketName, ok := marketNames[market]
	if !ok {
		return false
	}

	schedules, ok := marketSchedules[marketName]
	if !ok {
		return false
	}

	currentMinutes := now.Hour()*100 + now.Minute()
	for _, schedule := range schedules {
		if currentMinutes >= schedule.Open && currentMinutes < schedule.Close {
			return true
		}
	}

	return false
}

func GetMarketStatus(market string) string {
	loc, ok := marketTimezones[market]
	if !ok {
		return "未知市场"
	}

	now := time.Now().In(loc)
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return fmt.Sprintf("休市 (周末)")
	}

	marketName, ok := marketNames[market]
	if !ok {
		return "未知市场"
	}

	schedules, ok := marketSchedules[marketName]
	if !ok {
		return "未知市场"
	}

	currentMinutes := now.Hour()*100 + now.Minute()
	for _, schedule := range schedules {
		if currentMinutes < schedule.Open {
			return fmt.Sprintf("休市 (开盘时间: %02d:%02d)", schedule.Open/100, schedule.Open%100)
		}
		if currentMinutes >= schedule.Open && currentMinutes < schedule.Close {
			return fmt.Sprintf("交易中 (收盘时间: %02d:%02d)", schedule.Close/100, schedule.Close%100)
		}
	}

	lastSchedule := schedules[len(schedules)-1]
	return fmt.Sprintf("休市 (收盘时间: %02d:%02d)", lastSchedule.Close/100, lastSchedule.Close%100)
}

func GetMarketTime(market string) time.Time {
	loc, ok := marketTimezones[market]
	if !ok {
		return time.Now()
	}
	return time.Now().In(loc)
}
