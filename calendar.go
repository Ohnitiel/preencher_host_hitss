package main

import "time"

type Calendar map[time.Time]bool

func calculateEaster(year int) time.Time {
	Gauss := map[int][2]int{
		1599: {22, 2},
		1699: {22, 2},
		1799: {23, 3},
		1899: {24, 4},
		1999: {24, 5},
		2099: {24, 5},
		2199: {24, 6},
		2299: {25, 7},
	}
	century := (year/100)*100 + 99
	x := Gauss[century][0]
	y := Gauss[century][1]
	a := year % 19
	b := year % 4
	c := year % 7
	d := (19*a + x) % 30
	e := (2*b + 4*c + 6*d + y) % 7

	if (d + e) > 9 {
		day := d + e - 9
		if day == 26 {
			day = 19
		}
		if day == 25 && d == 28 && a > 10 {
			day = 18
		}
		return time.Date(year, time.April, day, 0, 0, 0, 0, time.UTC)
	} else {
		day := d + e + 22
		return time.Date(year, time.March, day, 0, 0, 0, 0, time.UTC)
	}
}

func CalendarForYear(year int) Calendar {
	var calendar Calendar

	current_year := time.Now().Year()
	pascoa := calculateEaster(current_year)
	carnival := pascoa.AddDate(0, 0, -47)
	corpus_christi := pascoa.AddDate(0, 0, 60)

	holidays := map[time.Time]bool{
		carnival:                 true,
		corpus_christi:           true,
		pascoa.AddDate(0, 0, -2): true, // Sexta-feira Santa
		time.Date(current_year, time.January, 1, 0, 0, 0, 0, time.UTC):   true, // Ano Novo
		time.Date(current_year, time.April, 21, 0, 0, 0, 0, time.UTC):    true, // Tiradentes
		time.Date(current_year, time.May, 1, 0, 0, 0, 0, time.UTC):       true, // Dia do Trabalho
		time.Date(current_year, time.September, 7, 0, 0, 0, 0, time.UTC): true, // Independência do Brasil
		time.Date(current_year, time.October, 12, 0, 0, 0, 0, time.UTC):  true, // Nossa Senhora Aparecida
		time.Date(current_year, time.November, 2, 0, 0, 0, 0, time.UTC):  true, // Finados
		time.Date(current_year, time.November, 15, 0, 0, 0, 0, time.UTC): true, // Proclamação da República
		time.Date(current_year, time.November, 21, 0, 0, 0, 0, time.UTC): true, // Consciência Negra
		time.Date(current_year, time.December, 25, 0, 0, 0, 0, time.UTC): true, // Natal
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		is_business_day := false
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday && !holidays[d] {
			is_business_day = true
		}
		calendar[d] = is_business_day
	}

	return calendar
}
