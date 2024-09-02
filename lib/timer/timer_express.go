package timer

import (
	"strconv"
	"strings"
	"time"
)

const (
	Max_WeekDay = 7  //max weekday value
	Min_WeekDay = 0  //min weekday value
	Max_Month   = 12 //max month value
	Min_Month   = 1  //min month value
	Max_Day     = 31 //max day value
	Min_Day     = 1  //min day value
	Max_Hour    = 23 //max hour value
	Min_Hour    = 0  //min hour value
	Max_Minute  = 59 //max minute value
	Min_Minute  = 0  //min minute value
	Max_Second  = 59 //max second value
	Min_Second  = 0  //min second value
)

const (
	ExpressType_WeekDay = "weekday"
	ExpressType_Month   = "month"
	ExpressType_Day     = "day"
	ExpressType_Hour    = "hour"
	ExpressType_Minute  = "minute"
	ExpressType_Second  = "second"
)

type ExpressSet struct {
	timeMap     map[int]int
	expressType string
	rawExpress  string
}

func (e *ExpressSet) IsMatch(t time.Time) bool {
	switch e.expressType {
	case ExpressType_WeekDay:
		_, ok := e.timeMap[int(t.Weekday())]
		return ok
	case ExpressType_Month:
		_, ok := e.timeMap[int(t.Month())]
		return ok
	case ExpressType_Day:
		_, ok := e.timeMap[int(t.Day())]
		return ok
	case ExpressType_Hour:
		_, ok := e.timeMap[int(t.Hour())]
		return ok
	case ExpressType_Minute:
		_, ok := e.timeMap[int(t.Minute())]
		return ok
	case ExpressType_Second:
		_, ok := e.timeMap[int(t.Second())]
		return ok
	default:
		return false
	}
}

/*
parse express
maybe：
"*" "*\/5" "1,2,3,5-8"
*/
func parseExpress(express, expressType string) (times *ExpressSet) {
	V_Max := 0
	V_Min := 0

	switch expressType {
	case ExpressType_WeekDay:
		V_Max = Max_WeekDay
		V_Min = Min_WeekDay
	case ExpressType_Month:
		V_Max = Max_Month
		V_Min = Min_Month
	case ExpressType_Day:
		V_Max = Max_Day
		V_Min = Min_Day
	case ExpressType_Hour:
		V_Max = Max_Hour
		V_Min = Min_Hour
	case ExpressType_Minute:
		V_Max = Max_Minute
		V_Min = Min_Minute
	case ExpressType_Second:
		V_Max = Max_Second
		V_Min = Min_Second
	default:
		return nil
	}

	times = &ExpressSet{timeMap: make(map[int]int), expressType: expressType, rawExpress: express}

	//if contains "/", reset frequency
	frequency := 1
	var err error
	if strings.Contains(express, "/") {
		frequency, err = strconv.Atoi(subString(express, strings.Index(express, "/")+1, -1))
		if err != nil {
			return nil
		}
		express = subString(express, 0, strings.Index(express, "/"))
	}

	//if express like "*" or "*/5"
	if express == "*" {
		for i := V_Min; i <= V_Max; i = i + frequency {
			times.timeMap[i] = i
		}
	} else {
		//if express like "1,2,4,12-20" or "12-2"
		//tops:if parse int failed, will be ignore
		tmpVals := strings.Split(express, ",")
		for i := 0; i < len(tmpVals); i++ {
			val := tmpVals[i]
			if val == "" {
				continue
			}
			//if not exists "-", parse int and join to times
			if !strings.Contains(val, "-") {
				atoi, errAtoi := strconv.Atoi(val)
				if errAtoi == nil {
					if _, ok := times.timeMap[atoi]; !ok {
						times.timeMap[atoi] = atoi
					}
				}
				continue
			}

			//if exists "-",parse max and min
			//tips:only deal first "-"
			firstString := subString(val, 0, strings.Index(val, "-"))
			lastString := subString(val, strings.Index(val, "-")+1, -1)
			var first, last int
			if first, err = strconv.Atoi(firstString); err != nil {
				//转义失败，忽略
				continue
			}
			if last, err = strconv.Atoi(lastString); err != nil {
				//转义失败，忽略
				continue
			}
			if last > V_Max || last < V_Min || first < V_Min || first > V_Max {
				//转义失败，忽略
				continue
			}

			//one loop, like 5-7
			if last >= first {
				for i := first; i <= last; i = i + frequency {
					if _, ok := times.timeMap[i]; !ok {
						times.timeMap[i] = i
					}
				}
			}
			//not one loop, like 12-2
			if last < first {
				for i := first; i <= V_Max; i = i + frequency {
					if _, ok := times.timeMap[i]; !ok {
						times.timeMap[i] = i
					}
				}
				for i := V_Min; i <= last; i = i + frequency {
					if _, ok := times.timeMap[i]; !ok {
						times.timeMap[i] = i
					}
				}
			}
		}

	}

	//if type is weekday, change 7 to 0
	if expressType == ExpressType_WeekDay {
		if _, ok := times.timeMap[7]; ok {
			delete(times.timeMap, 7)
			//modify for #8
			if _, ok := times.timeMap[0]; !ok {
				times.timeMap[0] = 0
			}
		}
	}
	return times

}

func subString(str string, begin, end int) string {
	rs := []rune(str)
	length := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= length {
		return ""
	}
	if end == -1 || end > length {
		return string(rs[begin:])
	}
	return string(rs[begin:end])
}
