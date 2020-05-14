package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (env *Environment) handleGetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{Message: "OK"})
}

func (env *Environment) handleGetMySQLVariables(c *gin.Context) {
	fqdn := c.Param("fqdn")
	if len(fqdn) == 0 {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "missing required parameter fqdn"})
		return
	}
	fqdnParts := strings.Split(fqdn, ":")
	host := fqdnParts[0]
	port := ""
	if len(fqdnParts) > 1 {
		port = fqdnParts[1]
	}

	variableNames := c.Query("variables")
	variables, err := GetMySQLVariables(env.DB.Username, env.DB.Password, host, port, "performance_schema", "", variableNames)
	if err != nil {
		c.JSON(http.StatusOK, APIResponse{Error: fmt.Sprintf("error getting variables: %s", err.Error())})
		return
	}
	variableJSON, err := json.Marshal(variables)
	if err != nil {
		c.JSON(http.StatusOK, APIResponse{Error: fmt.Sprintf("error marshaling JSON: %s", err.Error())})
	}
	c.JSON(http.StatusOK, APIResponse{Message: string(variableJSON)})
	return
}

func (env *Environment) handlePostMySQLVariables(c *gin.Context) {
	fqdn := c.Param("fqdn")
	if len(fqdn) == 0 {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "missing required parameter fqdn"})
		return
	}
	fqdnParts := strings.Split(fqdn, ":")
	host := fqdnParts[0]
	port := ""
	if len(fqdnParts) > 1 {
		port = fqdnParts[1]
	}

	// "variable_name=super_read_only&variable_value=on"
	// "variable_name=read_only&variable_value=off"
	varNames, ok := c.GetPostFormArray("variable_name")
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "missing required parameter(s) variable_name"})
		return
	}
	varValues, ok := c.GetPostFormArray("variable_value")
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "missing required parameter(s) variable_value"})
		return
	}
	if len(varNames) != len(varValues) {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "number of variable names and values must match"})
		return
	}

	// Create a map of variable_name to variable_values
	// Ex: {"super_read_only":"on", "read_only":"on"}
	variableSettings := make(map[string]string, len(varNames))
	for i, varName := range varNames {
		variableSettings[varName] = varValues[i]
	}

	variables, err := SetMySQLVariables(env.DB.Username, env.DB.Password, host, port, "performance_schema", "", variableSettings)
	if err != nil {
		c.JSON(http.StatusOK, APIResponse{Error: fmt.Sprintf("error setting variables: %s", err.Error())})
		return
	}

	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		c.JSON(http.StatusOK, APIResponse{Error: fmt.Sprintf("error marshaling JSON: %s", err.Error())})
	}
	c.JSON(http.StatusOK, APIResponse{Message: string(variablesJSON)})
	return
}
