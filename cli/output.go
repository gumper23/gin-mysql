package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"text/tabwriter"
)

// MySQLVariable variables have a name and a value
type MySQLVariable struct {
	Name  string `json:"variable_name"`
	Value string `json:"variable_value"`
}

// APIResponse is the standard response from the API
type APIResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (env *Environment) outputVars(getVarsVariables string, getVarsFQDNs []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	apiResponses := make(map[string]APIResponse)
	for _, fqdn := range getVarsFQDNs {
		wg.Add(1)
		go func(api, vars, fqdn string) {
			defer wg.Done()

			adminURL := fmt.Sprintf("http://%s/api/v1/mysql/variables/%s?variables=%s", api, fqdn, vars)
			resp, err := http.Get(adminURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "API call error: %s\n", err.Error())
				return
			}

			apiResponse := APIResponse{}
			err = json.NewDecoder(resp.Body).Decode(&apiResponse)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON decode error: %s\n", err.Error())
				return
			}

			mu.Lock()
			apiResponses[fqdn] = apiResponse
			mu.Unlock()
		}(env.ApiFQDN, getVarsVariables, fqdn)
	}
	wg.Wait()

	for _, fqdn := range getVarsFQDNs {
		if len(apiResponses[fqdn].Error) != 0 {
			fmt.Fprintf(os.Stderr, "Error from API: %s\n", apiResponses[fqdn].Error)
			continue
		}
		var mysqlVars []MySQLVariable
		err := json.Unmarshal([]byte(apiResponses[fqdn].Message), &mysqlVars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling JSON from %s: %s\n", fqdn, err.Error())
			continue
		}

		// Header
		fmt.Printf("Settings for %s\n", fqdn)
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
		fmt.Fprintf(w, "%s\t%s\t\n", "Name", "Value")
		for _, mysqlVar := range mysqlVars {
			fmt.Fprintf(w, "%s\t%s\t\n", mysqlVar.Name, mysqlVar.Value)
		}
		w.Flush()
		fmt.Println("")
	}
}
