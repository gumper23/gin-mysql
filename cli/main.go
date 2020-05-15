package main

import (
	"fmt"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("gin-mysql", "A CLI for gin-mysql-api")
	apiFQDN = app.Flag("api-fqdn", "The FQDN of the API server").Short('a').Default("localhost:8080").String()

	getVars          = app.Command("get-vars", "Retrieves a list of global variables")
	getVarsVariables = getVars.Flag("variables", "Ex: -v super_read_only,read_only").Short('v').String()
	getVarsFQDNs     = getVars.Arg("fqdns", "A space-separated list of FQDNs. Ex: 127.0.0.1:13306").Required().Strings()

	setVars          = app.Command("setvars", "Sets a list of global variables")
	setVarsVariables = setVars.Flag("variable", "Ex: -v super_read_only=1,read_only=1").Short('v').Required().Strings()
	setVarsFQDNs     = setVars.Arg("fqdns", "A space-separated list of FQDNs. Ex: 127.0.0.1:13306").Required().Strings()
)

// Environment holds environment settings
type Environment struct {
	ApiFQDN string `toml:"api"`
}

func main() {
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	env := &Environment{ApiFQDN: *apiFQDN}
	switch cmd {
	// ./gin-mysql get-vars -v super_read_only,read_only 127.0.0.1:13306 127.0.0.1:23306 127.0.0.1:33306 127.0.0.1:43306
	// ./gin-mysql get-vars -v super_read_only,read_only,log_bin,server_uuid 127.0.0.1:13306 127.0.0.1:23306 127.0.0.1:33306 127.0.0.1:43306
	case getVars.FullCommand():
		env.outputGetVariables(*getVarsVariables, *getVarsFQDNs)
	case setVars.FullCommand():
		fmt.Printf("%s\n", setVars.FullCommand())
	}
}
