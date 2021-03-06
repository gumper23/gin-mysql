package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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

func (env *Environment) outputGetQueries(fqdns []string) {
	var wg sync.WaitGroup
	var outMu sync.Mutex
	var mapMu sync.Mutex

	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range fqdns {
		wg.Add(1)
		go func(api, fqdn string) {
			defer wg.Done()

			adminURL := fmt.Sprintf("http://%s/api/v1/mysql/queries/%s", api, fqdn)
			c := http.Client{
				Timeout: 3 * time.Second,
			}
			defer c.CloseIdleConnections()

			resp, err := c.Get(adminURL)
			if err != nil {
				outMu.Lock()
				w := ansiterm.NewWriter(os.Stderr)
				w.SetForeground(ansiterm.BrightRed)
				w.SetStyle(ansiterm.Bold)
				fmt.Fprintf(w, "error calling API on %s: %s\n", fqdn, err.Error())
				w.Reset()
				outMu.Unlock()
				return
			}
			defer resp.Body.Close()

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				outMu.Lock()
				w := ansiterm.NewWriter(os.Stderr)
				w.SetForeground(ansiterm.BrightRed)
				w.SetStyle(ansiterm.Bold)
				fmt.Fprintf(w, "error decoding JSON on %s: %s\n", fqdn, err.Error())
				w.Reset()
				outMu.Unlock()
				return
			}

			mapMu.Lock()
			apiResponses[fqdn] = apiResponse
			mapMu.Unlock()

		}(env.ApiFQDN, fqdn)
	}
	wg.Wait()

	w := ansiterm.NewTabWriter(os.Stdout, 1, 1, 4, ' ', 0)
	w.SetForeground(ansiterm.BrightBlue)
	w.SetStyle(ansiterm.Bold)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "FQDN", "Time", "Microseconds", "Queries")
	w.Reset()
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
			queries := make(map[string]string, 0)
			err := json.Unmarshal([]byte(apiResponses[fqdn].Message), &queries)
			if err != nil {
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(os.Stderr, "Error unmarshalling JSON on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				continue
			}

			// Body
			w.SetForeground(ansiterm.BrightGreen)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", fqdn, queries["queries_ts"], queries["queries_us"], queries["queries"])
			w.Reset()
		}
	}
	w.Flush()
}

func (env *Environment) outputGetVariables(variables string, fqdns []string) {
	var wg sync.WaitGroup
	var mapMu sync.Mutex
	var outMu sync.Mutex

	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range fqdns {
		wg.Add(1)
		go func(api, vars, fqdn string) {
			defer wg.Done()

			adminURL := fmt.Sprintf("http://%s/api/v1/mysql/variables/%s", api, fqdn)
			if len(variables) > 0 {
				adminURL = fmt.Sprintf("%s?variables=%s", adminURL, vars)
			}
			c := http.Client{
				Timeout: 3 * time.Second,
			}
			defer c.CloseIdleConnections()

			resp, err := c.Get(adminURL)
			if err != nil {
				outMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "API call error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				outMu.Unlock()
				return
			}

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				outMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "JSON decode error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				outMu.Unlock()
				return
			}

			mapMu.Lock()
			apiResponses[fqdn] = apiResponse
			mapMu.Unlock()

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

// super_read_only=1,read_only=1
func (env *Environment) outputSetVariables(settings string, fqdns []string) {
	var wg sync.WaitGroup
	var mapMu sync.Mutex
	var outMu sync.Mutex

	if len(settings) == 0 {
		fmt.Fprintf(os.Stderr, "settings empty - nothing to set\n")
		return
	}
	// Convert the settings string "var1=val1,var2=val2,varN=valN into a map[string]string"
	settingsSlice := strings.Split(settings, ",")
	settingsMap := make(map[string]string, len(settingsSlice))
	for _, setting := range settingsSlice {
		settingParts := strings.Split(setting, "=")
		if len(settingParts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid setting %s; no '=' sign", setting)
			return
		}
		settingsMap[settingParts[0]] = settingParts[1]
	}
	settingsJSON, err := json.Marshal(settingsMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling settings map to JSON: %s\n", err.Error())
		return
	}

	// gin-mysql set-vars super_read_only=1,read_only=1 127.0.0.1:23306 127.0.0.1:33306 127.0.0.1:43306
	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range fqdns {
		wg.Add(1)
		go func(api, vars, fqdn string) {
			defer wg.Done()

			apiURL := fmt.Sprintf("http://%s/api/v1/mysql/variables/%s", api, fqdn)
			c := http.Client{
				Timeout: 3 * time.Second,
			}
			resp, err := c.Post(apiURL, "application/json", bytes.NewBuffer(settingsJSON))
			if err != nil {
				outMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "API call error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				outMu.Unlock()
				return
			}

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				outMu.Lock()
				werr := ansiterm.NewWriter(os.Stderr)
				werr.SetForeground(ansiterm.BrightRed)
				werr.SetStyle(ansiterm.Bold)
				fmt.Fprintf(werr, "JSON decode error on %s: %s\n", fqdn, err.Error())
				werr.Reset()
				outMu.Unlock()
				return
			}

			mapMu.Lock()
			apiResponses[fqdn] = apiResponse
			mapMu.Unlock()

		}(env.ApiFQDN, settings, fqdn)
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
