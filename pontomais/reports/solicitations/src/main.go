package main

import (
	"formula/pkg/formula"
	"os"
	"time"
)

func main() {
	startDate := os.Getenv("RIT_START_DATE")
	endDate := os.Getenv("RIT_END_DATE")
	period := os.Getenv("RIT_PERIOD")
	username := os.Getenv("RIT_PONTOMAIS_LOGIN")
	password := os.Getenv("RIT_PONTOMAIS_PASSWORD")

	if period != "Outros" {
		now := time.Now()
		endDate = now.String()[:10]
		switch period {
		case "Ultima semana":
			startDate = now.AddDate(0, 0, -7).String()[:10]
		case "Ultima quinzena":
			startDate = now.AddDate(0, 0, -15).String()[:10]
		case "Ultimo mes":
			startDate = now.AddDate(0, -1, 0).String()[:10]
		}
	}

	formula.Formula{
		StartDate: startDate,
		EndDate:   endDate,
		Username:  username,
		Password:  password,
	}.Run()
}
