package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response from the API
type APIResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (env *Environment) handleGetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{Message: "OK"})
}

func (env *Environment) handleGetMySQLQueries(c *gin.Context) {
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

	queries, err := GetMySQLQueries(env.DB.Username, env.DB.Password, host, port, "performance_schema", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Error: fmt.Sprintf("error getting status: %s", err.Error())})
		return
	}
	queriesJSON, err := json.Marshal(queries)
	if err != nil {
		c.JSON(http.StatusOK, APIResponse{Error: fmt.Sprintf("error marshaling JSON: %s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Message: string(queriesJSON)})
	return
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
		return
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

	settings := make(map[string]string)
	err := c.BindJSON(&settings)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "missing settings JSON"})
		return
	}
	if len(settings) == 0 {
		c.JSON(http.StatusUnprocessableEntity, APIResponse{Error: "empty settings JSON"})
		return
	}

	variables, err := SetMySQLVariables(env.DB.Username, env.DB.Password, host, port, "performance_schema", "", settings)
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
