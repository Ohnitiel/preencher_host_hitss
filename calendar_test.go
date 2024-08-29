package main

import (
	"testing"
	"time"
)

func TestVariatingHolidays(t *testing.T) {
	var allHolidays []time.Time

	variatingHolidays := []time.Time{
		time.Date(2015, time.February, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.April, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.June, 4, 0, 0, 0, 0, time.UTC),

		time.Date(2016, time.February, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2016, time.March, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2016, time.May, 26, 0, 0, 0, 0, time.UTC),

		time.Date(2017, time.February, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.April, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2017, time.June, 15, 0, 0, 0, 0, time.UTC),

		time.Date(2018, time.February, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2018, time.March, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2018, time.May, 31, 0, 0, 0, 0, time.UTC),

		time.Date(2019, time.March, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.April, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.June, 20, 0, 0, 0, 0, time.UTC),

		time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.June, 11, 0, 0, 0, 0, time.UTC),

		time.Date(2021, time.February, 16, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.June, 3, 0, 0, 0, 0, time.UTC),

		time.Date(2022, time.March, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.April, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.June, 16, 0, 0, 0, 0, time.UTC),

		time.Date(2023, time.February, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2023, time.April, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2023, time.June, 8, 0, 0, 0, 0, time.UTC),

		time.Date(2024, time.February, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.May, 30, 0, 0, 0, 0, time.UTC),
	}

	for year := 2015; year <= 2024; year++ {
		holidays := CalendarForYear(year)
		for d, v := range holidays {
			if v == false {
				allHolidays = append(allHolidays, d)
			}
		}
	}

	present := false
	for _, v := range variatingHolidays {
		for _, d := range allHolidays {
			if v == d {
				present = true
			}
		}
		if !present {
			t.Errorf("Variating holiday %v not found", v)
		}
		present = false
	}
}
