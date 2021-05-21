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

	"github.com/jedib0t/go-pretty/table"
)

const ApiVersion = "2"

type auth struct {
	clientId string
	token    string
	email    string
}

type Formula struct {
	StartDate string
	EndDate   string
	Username  string
	Password  string
	Client    http.Client
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
			"columns":    "date,updated_at,is_medical_certificate,status,employee_name,observation,time_cards,answered_by,solicitation_status,solicitation_type",
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
		"https://api.pontomais.com.br/api/html_reports/solicitations",
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

func (f Formula) showReport(records [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Nome",
		"Data",
		"Status",
		"Tipo solicitação",
		"Data da alteração",
		"Observação",
		"Motivo",
		"É atestado?",
	})

	ok := false
	statusColumn := 999
	for _, record := range records {
	    if statusColumn == 999 {
            for j, columns := range record {
                if columns == "Status" {
                    statusColumn = j
                    break
                }
            }
	    }
		if record[0] == "Nome" {
			ok = true
			continue
		} else if record[0] == "Resumo" {
			ok = false
		}
		if ok && record[statusColumn] == "Pendente" {
			t.AppendRow(table.Row{
				record[0],
				record[1],
				record[statusColumn],
				record[statusColumn+1],
				record[statusColumn+3],
				record[statusColumn+4],
				record[statusColumn+5],
				record[statusColumn+6],
			})
		}
	}

	t.SetStyle(table.StyleLight)
	t.Render()
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

	f.showReport(records)
}
