package timepicker

import (
	"strconv"
	"strings"
)

const (
	//JamPertama timepicker
	JamPertama = "09"
	//MenitPertama timepicker
	MenitPertama = "00"
	//JamTerakhir timepicker
	JamTerakhir = "17"
	//MenitTerakhir timepicker
	MenitTerakhir = "59"
)

func CreateCallbackData(action string, jam string, menit string) string {
	return strings.Join([]string{action, jam, menit}, ";")
}

func SeparateCallbackData(data string) []string {
	return strings.Split(data, ";")
}

func CreateNextJam(data string) string {
	i, _ := strconv.Atoi(data)

	next := i + 1

	jamAkhir, _ := strconv.Atoi(JamTerakhir)
	jamAwal, _ := strconv.Atoi(JamPertama)

	if next > jamAkhir {
		next = jamAwal
	}

	j := strconv.Itoa(next)
	if next < 10 {
		j = "0" + j
	}
	return j
}

func CreatePrevJam(data string) string {
	i, _ := strconv.Atoi(data)

	prev := i - 1

	jamAkhir, _ := strconv.Atoi(JamTerakhir)
	jamAwal, _ := strconv.Atoi(JamPertama)

	if prev < jamAwal {
		prev = jamAkhir
	}

	j := strconv.Itoa(prev)
	if prev < 10 {
		j = "0" + j
	}
	return j
}

func CreateNextMenit(data string) string {
	i, _ := strconv.Atoi(data)

	next := i + 5

	menitAkhir, _ := strconv.Atoi(MenitTerakhir)

	if next > menitAkhir {
		next = 0
	}

	m := strconv.Itoa(next)
	if next < 10 {
		m = "0" + m
	}
	return m
}

func CreatePrevMenit(data string) string {
	i, _ := strconv.Atoi(data)

	prev := i - 5

	menitAwal, _ := strconv.Atoi(MenitPertama)

	if prev < menitAwal {
		prev = 55
	}

	m := strconv.Itoa(prev)
	if prev < 10 {
		m = "0" + m
	}
	return m
}
