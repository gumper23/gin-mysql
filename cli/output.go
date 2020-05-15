package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/juju/ansiterm"
)

// APIResponse is the standard response from the API
type APIResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// MySQLVariable variables have a name and a value
type MySQLVariable struct {
	Name  string `json:"variable_name"`
	Value string `json:"variable_value"`
}

func (env *Environment) outputGetVariables(variables string, fqdns []string) {
	var wg sync.WaitGroup
	var storeMu sync.Mutex
	var writeMu sync.Mutex

	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range fqdns {
		wg.Add(1)
		go func(api, vars, fqdn string) {
			defer wg.Done()

			adminURL := fmt.Sprintf("http://%s/api/v1/mysql/variables/%s", api, fqdn)
			if len(variables) > 0 {
				adminURL = fmt.Sprintf("%s?variables=%s", adminURL, vars)
			}
			resp, err := http.Get(adminURL)
			if err != nil {
				writeMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "API call error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				writeMu.Unlock()
				return
			}

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				writeMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "JSON decode error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				writeMu.Unlock()
				return
			}

			storeMu.Lock()
			apiResponses[fqdn] = apiResponse
			storeMu.Unlock()

		}(env.ApiFQDN, variables, fqdn)
	}
	wg.Wait()

	for _, fqdn := range fqdns {
		if len(apiResponses[fqdn].Error) != 0 {
			werr := ansiterm.NewWriter(os.Stderr)
			werr.SetForeground(ansiterm.BrightRed)
			werr.SetStyle(ansiterm.Bold)
			fmt.Fprintf(os.Stderr, "Error from API: %s\n", apiResponses[fqdn].Error)
			werr.Reset()
			continue
		}

		if len(apiResponses[fqdn].Message) > 0 {
			var mysqlVars []MySQLVariable
			err := json.Unmarshal([]byte(apiResponses[fqdn].Message), &mysqlVars)
			if err != nil {
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(os.Stderr, "Error unmarshaling JSON from %s: %s\n", fqdn, err.Error())
				werr.Reset()
				continue
			}

			// Header
			wout := ansiterm.NewWriter(os.Stdout)
			wout.SetStyle(ansiterm.Bold)
			fmt.Fprintf(wout, "Settings for %s:\n", fqdn)
			wout.Reset()

			w := ansiterm.NewTabWriter(os.Stdout, 1, 1, 4, ' ', 0)
			fmt.Fprintf(w, "%s\t%s\t\n", "Name", "Value")
			for _, mysqlVar := range mysqlVars {
				fmt.Fprintf(w, "%s\t%s\t\n", mysqlVar.Name, mysqlVar.Value)
			}
			w.Flush()
			w.Reset()

			fmt.Println("")
		}
	}
}

// -v super_read_only=1,read_only=1
func (env *Environment) outputSetVariables(variables string, fqdns []string) {
	var wg sync.WaitGroup
	var storeMu sync.Mutex
	var writeMu sync.Mutex

	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range fqdns {
		wg.Add(1)
		go func(api, vars, fqdn string) {
			defer wg.Done()

			adminURL := fmt.Sprintf("http://%s/api/v1/mysql/variables/%s", api, fqdn)
			if len(variables) > 0 {
				adminURL = fmt.Sprintf("%s?variables=%s", adminURL, vars)
			}
			resp, err := http.Get(adminURL)
			if err != nil {
				writeMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "API call error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				writeMu.Unlock()
				return
			}

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				writeMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "JSON decode error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				writeMu.Unlock()
				return
			}

			storeMu.Lock()
			apiResponses[fqdn] = apiResponse
			storeMu.Unlock()

		}(env.ApiFQDN, variables, fqdn)
	}
	wg.Wait()

	for _, fqdn := range fqdns {
		if len(apiResponses[fqdn].Error) != 0 {
			werr := ansiterm.NewWriter(os.Stderr)
			werr.SetForeground(ansiterm.BrightRed)
			werr.SetStyle(ansiterm.Bold)
			fmt.Fprintf(os.Stderr, "Error from API: %s\n", apiResponses[fqdn].Error)
			werr.Reset()
			continue
		}

		if len(apiResponses[fqdn].Message) > 0 {
			var mysqlVars []MySQLVariable
			err := json.Unmarshal([]byte(apiResponses[fqdn].Message), &mysqlVars)
			if err != nil {
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(os.Stderr, "Error unmarshaling JSON from %s: %s\n", fqdn, err.Error())
				werr.Reset()
				continue
			}

			// Header
			wout := ansiterm.NewWriter(os.Stdout)
			wout.SetStyle(ansiterm.Bold)
			fmt.Fprintf(wout, "Settings for %s:\n", fqdn)
			wout.Reset()

			w := ansiterm.NewTabWriter(os.Stdout, 1, 1, 4, ' ', 0)
			fmt.Fprintf(w, "%s\t%s\t\n", "Name", "Value")
			for _, mysqlVar := range mysqlVars {
				fmt.Fprintf(w, "%s\t%s\t\n", mysqlVar.Name, mysqlVar.Value)
			}
			w.Flush()
			w.Reset()

			fmt.Println("")
		}
	}
}
