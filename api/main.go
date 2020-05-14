package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

// Listener holds listener configuration information
type Listener struct {
	Port string `toml:"port"`
}

// Database holds database credentials
type Database struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type tomlConfig struct {
	Listen Listener `toml:"listener"`
	DB     Database `toml:"database"`
}

// Environment holds the environmental configuration
type Environment struct {
	Port string   `toml:"port"`
	DB   Database `toml:"database"`
}

// APIResponse is the standard response from the API
type APIResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func main() {
	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatal(err)
		return
	}

	env := &Environment{
		Port: config.Listen.Port,
		DB: Database{
			Username: config.DB.Username,
			Password: config.DB.Password,
		},
	}

	r := gin.Default()

	// curl http://localhost:8080/ | jq .
	// curl http://localhost:8080/status | jq .
	r.GET("/", env.handleGetStatus)
	r.GET("/status", env.handleGetStatus)

	// curl http://localhost:8080/api/v1/mysql/variables/127.0.0.1:43306 | jq '.message | fromjson'
	// curl http://localhost:8080/api/v1/mysql/variables/127.0.0.1:43306?variables=super_read_only,read_only | jq '.message | fromjson'
	r.GET("/api/v1/mysql/variables/:fqdn", env.handleGetMySQLVariables)

	// curl -d 'variable_name=super_read_only&variable_value=off' http://localhost:8080/api/v1/mysql/variables/127.0.0.1:43306 | jq '.message | fromjson'
	// curl -d 'variable_name=super_read_only&variable_value=off' -d 'variable_name=read_only&variable_value=on' http://localhost:8080/api/v1/mysql/variables/127.0.0.1:43306 | jq '.message | fromjson'
	r.POST("/api/v1/mysql/variables/:fqdn", env.handleSetMySQLVariables)

	fmt.Printf("Listening on port %s...\n", env.Port)
	r.Run(":" + env.Port)
}
