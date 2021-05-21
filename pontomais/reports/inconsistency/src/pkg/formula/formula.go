package formula

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

const ApiVersion = "2"

type InconsistentDate struct {
	Date    string
	Records []string
	Reason  string
}

type Inconsistency struct {
	Collaborator     string
	InconsistentDate []InconsistentDate
}

type auth struct {
	clientId string
	token    string
	email    string
}

type Formula struct {
	MinRecords int
	StartDate  string
	EndDate    string
	Username   string
	Password   string
	Client     http.Client
}

func (f Formula) login() (auth, error) {
	body, err := json.Marshal(map[string]string{
		"login":    f.Username,
		"password": f.Password,
	})
	if err != nil {
		return auth{}, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		"https://api.pontomais.com.br/api/auth/sign_in",
		bytes.NewReader(body),
	)
	if err != nil {
		return auth{}, err
	}

	request.Header.Add("api-version", ApiVersion)
	request.Header.Add("authority", "api.pontomais.com.br")
	request.Header.Add("content-type", "application/json;charset=UTF-8")
	request.Header.Add("accept", "application/json, text/plain, */*")

	fmt.Println("🔓 Realizando login")
	resp, err := f.Client.Do(request)
	if err != nil {
		return auth{}, err
	}

	if resp.StatusCode <= 199 || resp.StatusCode >= 300 {
		return auth{}, fmt.Errorf("Failed to login, status code: %d", resp.StatusCode)
	}

	var rBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rBody)
	if err != nil {
		return auth{}, err
	}

	return auth{
		token:    fmt.Sprintf("%s", rBody["token"]),
		clientId: fmt.Sprintf("%s", rBody["client_id"]),
		email:    fmt.Sprintf("%s", reflect.ValueOf(rBody["data"]).MapIndex(reflect.ValueOf("email"))),
	}, nil
}

func (f Formula) getCSVReport(auth auth) ([][]string, error) {
	body := map[string]interface{}{
		"report": map[string]string{
			"columns":    "overnight_time,date,motive,time_cards,time_balance,summary,extra_time",
			"start_date": f.StartDate,
			"end_date":   f.EndDate,
			"format":     "csv",
		},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		"https://api.pontomais.com.br/api/html_reports/work_days",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Add("authority", "api.pontomais.com.br")
	request.Header.Add("content-type", "application/json;charset=UTF-8")
	request.Header.Add("accept", "application/json, text/plain, */*")
	request.Header.Add("api-version", ApiVersion)
	request.Header.Add("token-type", "Bearer")
	request.Header.Add("access-token", auth.token)
	request.Header.Add("client", auth.clientId)
	request.Header.Add("uid", auth.email)

	res, err := f.Client.Do(request)

	if err != nil {
		return nil, err
	}

	if res.StatusCode <= 199 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Failed to get csv, status code: %d", res.StatusCode)
	}

	var resBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&resBody)
	if err != nil {
		return nil, err
	}

	resCSV, err := http.Get(fmt.Sprintf("%s", resBody["url"]))
	if err != nil {
		return nil, err
	}

	defer resCSV.Body.Close()

	b, err := ioutil.ReadAll(resCSV.Body)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(b))
	r.FieldsPerRecord = -1

	return r.ReadAll()
}

func (f Formula) isInconsistent(dayRecords []string, date string, obs string) string {
	var count int

	for _, record := range dayRecords {
		if len(record) > 0 {
			count++
		}
	}

	if count == 0 {
		if !strings.Contains(date, "Sáb") && !strings.Contains(date, "Dom") && len(obs) == 0 {
			return "Falta"
		}
	} else if count < f.MinRecords {
		return "Menos registros"
	} else if count > f.MinRecords {
		return "Mais  registros"
	}

	return ""
}

func (f Formula) printResult(inconsistencies []Inconsistency, tableHeader []string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headerRow := table.Row{"Nome"}
	for _, a := range tableHeader {
		a = strings.Replace(a, "Entrada", "Ent.", 1)
		a = strings.Replace(a, "Motivo/Observação", "Obs.", 1)
		headerRow = append(headerRow, a)
	}
	headerRow = append(headerRow, "Inconsistência")
	t.AppendHeader(headerRow)

	for _, inconsistency := range inconsistencies {
		for _, inconsistentDate := range inconsistency.InconsistentDate {
			recordsRow := table.Row{inconsistency.Collaborator, inconsistentDate.Date}
			for _, record := range inconsistentDate.Records {
				recordsRow = append(recordsRow, record)
			}
			recordsRow = append(recordsRow, inconsistentDate.Reason)
			t.AppendRow(recordsRow)
		}
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

func (f Formula) showInconsistencies(records [][]string) {
	var inconsistencies []Inconsistency
	var inconsistentDates []InconsistentDate
	var collaborator string
	var tableHeader []string
	var creditIndex int

	for _, row := range records {
		if row[0] == "Colaborador" {
			collaborator = row[1]
			continue
		} else if row[0] == "TOTAIS" {
			if inconsistentDates != nil {
				inconsistencies = append(inconsistencies, Inconsistency{
					Collaborator:     collaborator,
					InconsistentDate: inconsistentDates,
				})

				inconsistentDates = nil
			}
			collaborator = ""
			continue
		} else if row[0] == "Data" {
			if tableHeader == nil {
				for i, elem := range row {
					if elem == "Crédito" {
						creditIndex = i
						break
					}
				}
				tableHeader = append(row[:creditIndex], row[len(row)-1])
			}
			continue
		}

		if len(collaborator) > 0 {
			inconsistency := f.isInconsistent(row[1:creditIndex], row[0], row[len(row)-1])
			if len(inconsistency) > 0 {
				inconsistentDates = append(inconsistentDates, InconsistentDate{
					Date:    row[0],
					Records: append(row[1:creditIndex], row[len(row)-1]),
					Reason:  inconsistency,
				})
			}
		}
	}

	f.printResult(inconsistencies, tableHeader)
}

func (f Formula) Run() {
	auth, err := f.login()
	if err != nil {
		log.Fatal(err)
	}

	records, err := f.getCSVReport(auth)
	if err != nil {
		log.Fatal(err)
	}

	f.showInconsistencies(records)
}
