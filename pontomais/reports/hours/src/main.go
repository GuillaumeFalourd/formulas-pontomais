package main

import (
	"formula/pkg/formula"
	"os"
	"time"
)

func main() {
	username := os.Getenv("RIT_PONTOMAIS_LOGIN")
	password := os.Getenv("RIT_PONTOMAIS_PASSWORD")

	now := time.Now()
	endDate := now.String()[:10]
	startDate := now.AddDate(0, 0, -1).String()[:10]

	formula.Formula{
		StartDate: startDate,
		EndDate:   endDate,
		Username:  username,
		Password:  password,
	}.Run()
}
