package main

import (
	"fmt"
	"time"
)

func main() {

	nativeTime := time.Date(2015, time.January, 1, 0, 0, 1, 9999, time.UTC)

	nativeTimeBytes, err := nativeTime.MarshalJSON()

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("native time : ", string(nativeTimeBytes))

	fixedTime := ISO8601Time(nativeTime)

	fixedTimeBytes, err := fixedTime.MarshalJSON()

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("fixed format : ", string(fixedTimeBytes))

	fixedTime2 := ISO8601Time{}

	err = fixedTime2.UnmarshalJSON(fixedTimeBytes)

	if err != nil {
		panic(err.Error())
	}

	fixedTimeBytes2, err := fixedTime2.MarshalJSON()

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("fixed format2 :", string(fixedTimeBytes2))
}

var ISO8601 = "2006-01-02T15:04:05.999Z07:00"

type ISO8601Time time.Time

func (t ISO8601Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(ISO8601)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, ISO8601)
	b = append(b, '"')
	return b, nil
}

func (t *ISO8601Time) UnmarshalJSON(data []byte) (err error) {
	tt, err := time.Parse(`"`+ISO8601+`"`, string(data))
	*t = ISO8601Time(tt)
	return
}
