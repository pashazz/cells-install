package main

import (
	"context"
	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/config"
	"github.com/pydio/cells/common/proto/install"
	"github.com/pydio/cells/common/utils/net"
	"github.com/pydio/cells/discovery/install/lib"
	"log"
	"net/url"
	"os"
	"syscall"
)

//This program installs Pydio Cells in automatic mode.
// Arguments are passed via environment variables
//Log is printed to stdin, as Docker expects

func printConfig (c *install.InstallConfig) {
	log.Printf("InternalUrl: %v\n", c.InternalUrl)
	log.Printf("DbConnectionType: %v\n", c.DbConnectionType)
	log.Printf("DBTCPHostname: %v\n", c.DbTCPHostname)
	log.Printf("DBTCPPort: %v\n", c.DbTCPPort)
	log.Printf("DBTCPName: %v\n", c.DbTCPName)
	log.Printf("DBTCPUser: %v\n", c.DbTCPUser)
	log.Printf("DBTcpPassword: %v\n", c.DbTCPPassword)
	log.Printf("DBSocketFile: %v\n", c.DbSocketFile)
	log.Printf("DBSocketName: %v\n", c.DbSocketName)
	log.Printf("DBSocketUser: %v\n", c.DbSocketUser)
	log.Printf("DBSocketPassword: %v\n", c.DbSocketPassword)
	log.Printf("DBManualDSN: %v\n", c.DbManualDSN)
	log.Printf("DsName: %v\n", c.DsName)
	log.Printf("DsPort: %v\n", c.DsPort)
	log.Printf("DsFolder: %v\n", c.DsFolder)
	log.Printf("ExternalMicro: %v\n", c.ExternalMicro)
	log.Printf("ExternalGateway: %v\n", c.ExternalGateway)
	log.Printf("ExternalWbesocket: %v\n", c.ExternalWebsocket)
	log.Printf("ExternalFrontPlugins: %v\n", c.ExternalFrontPlugins)
	log.Printf("ExternalDAV: %v\n", c.ExternalDAV)
	log.Printf("ExternalWOPI: %v\n", c.ExternalWOPI)
	log.Printf("ExternalDex: %v\n", c.ExternalDexID)
	log.Printf("ExternalDexSecret: %v\n", c.ExternalDexSecret)
	log.Printf("FrontendHosts: %v\n", c.FrontendHosts)
	log.Printf("FrontendLogin: %v\n", c.FrontendLogin)
	log.Printf("FrontendPassword: %v\n", c.FrontendPassword)
	log.Printf("FrontendRepeatPassword: %v\n", c.FrontendRepeatPassword)
	log.Printf("FrontendApplicationTitle: %v\n", c.FrontendApplicationTitle)
	log.Printf("FrontendDefaultLanguage: %v\n", c.FrontendDefaultLanguage)
	log.Printf("LicenseRequired: %v\n", c.LicenseRequired)
	log.Printf("LicenseString: %v\n", c.LicenseString)

	results := c.CheckResults
	for i := 0; i < len(results); i++ {
		log.Printf("CheckResult #%v:  '%v':  %v (%v)", i, results[i].Name, results[i].Success, results[i].JsonResult)
	}
}

func execveCells() {
	log.Printf("execve cells...")
	err := syscall.Exec("cells", []string{"cells", "start"}, os.Environ())
	if err != nil {
		log.Fatalf("Unable to execute cells: %v\n", err)
	}
}
func main() {
	//Check if FILE is available
	fileName := os.Getenv("FILE")
	if fileName == "" {
		log.Fatalf("Won't run as FILE environment variable is not present")
	}
	_, err := os.Stat(fileName)
	if !os.IsNotExist(err) {
		log.Printf("Running cells")
		execveCells()
		return
	} else {
		log.Printf("Running install..\n")
	}

	// Set go-micro ports
	micro := config.Get("ports", common.SERVICE_MICRO_API).Int(0)
		if micro == 0 {
			micro = net.GetAvailablePort()
			config.Set(micro, "ports", common.SERVICE_MICRO_API)
			config.Save("cli", "Install / Setting default Ports")
		}
	
	installConfig := lib.GenerateDefaultConfig()
	//EXTERNAL_URL is the url used to access application from the outside world, i.e. http://cells.com:8080
	externalUrl := os.Getenv("EXTERNAL_URL")
	_, e := url.Parse(externalUrl)
	if e != nil {
		log.Fatalf("Unable to parse 'EXTERNAL_URL' environment variable '%v' as valid URL: %v", externalUrl, e)
	}
	config.Set(externalUrl, "defaults", "url") //But EXTERNAL_URL does not participate in install config though

	//INTERNAL_URL is the address that a web server listens,, i.e. http://0.0.0.0:8080
	internalUrl := os.Getenv("INTERNAL_URL")
	_, err = url.Parse(internalUrl)
	if err != nil {
		log.Fatalf("Unable to parse 'INTERNAL_URL' environment variable '%v' as valid URL: %v", externalUrl, e)
	}
	if internalUrl == "" {
		log.Fatalf("INTERNAL_URL environment variable is required to be a server listening address, e.g. 0.0.0.0:8080")
	}
	installConfig.InternalUrl = internalUrl
	config.Set(internalUrl, "defaults", "urlInternal")


	//For now we don't configure SSL via this installer

	err = config.Save("cli", "Install / Setting default URLs")
	if err != nil {
		log.Fatalf("Unable to save config: %v", err)
	}


	DBConnError := func (env string, t string) {
			log.Fatalf("'%v' should not be empty if DB_CONNECTION_TYPE is %v\n", env, t)
		}
	//DB_CONNECTION_TYPE is either 'tcp' or 'socket' depending on what mysql connection type you want to use. If you want to input
	//DSN manually, leave this variable empty and set DB_DSN instead
	dbConnType := os.Getenv("DB_CONNECTION_TYPE")
	dbName := ""
	dbUser := ""
	dbPassword := ""
	if dbConnType == "tcp" || dbConnType == "socket" {
		dbUser = os.Getenv("DB_USER")
		dbPassword = os.Getenv("DB_PASSWORD")
		dbName = os.Getenv("DB_NAME")
	}
	switch dbConnType {
	case "tcp": {
			installConfig.DbConnectionType = "tcp"
			dbTcpHost := os.Getenv("DB_TCP_HOST")
			if dbTcpHost == "" {
				DBConnError("DB_TCP_HOST", "tcp")
			}
			installConfig.DbTCPHostname = dbTcpHost

			dbTcpPort := os.Getenv("DB_TCP_PORT")
			if dbTcpPort == "" {
				DBConnError("DB_TCP_PORT", "tcp")
			}
			installConfig.DbTCPPort = dbTcpPort

			if dbUser == "" {
				DBConnError("DB_USER", "tcp")
			}
			installConfig.DbTCPUser = dbUser

			if dbPassword == "" {
				DBConnError("DB_PASSWORD", "tcp")
			}
			installConfig.DbTCPPassword = dbPassword

			if dbName == "" {
				DBConnError("DB_NAME", "tcp")
			}
			installConfig.DbTCPName = dbName
		}
	case "socket": {
		dbSocket := os.Getenv("DB_SOCKET_FILE")
		if dbSocket == "" {
			DBConnError("DB_SOCKET_FILE", "socket")
		}
		installConfig.DbSocketFile = dbSocket

		if dbUser == "" {
			DBConnError("DB_USER", "socket")
		}
		installConfig.DbSocketUser = dbUser

		if dbPassword == "" {
			DBConnError("DB_PASSWORD", "socket")
		}
		installConfig.DbSocketPassword = dbPassword

		if dbName == "" {
			DBConnError("DB_NAME", "socket")
		}
		installConfig.DbSocketName = dbName
	}
	case "":
		dbDSN := os.Getenv("DB_DSN")
		if dbDSN == "" {
			DBConnError("DB_DSN", "not set")
		}
		installConfig.DbManualDSN =  dbDSN
	default:
		log.Fatalf("Failed to parse environment variable DB_CONNECTION_TYPE: expected: tcp, socket; got: %v\n", dbConnType)
	}
	if res := lib.PerformCheck(context.Background(), "DB", (*install.InstallConfig)(installConfig)); !res.Success {
		log.Fatalf("Can't connect to the database, check your environment variables:\n %v\n", res.GetJsonResult())
	} else {
		log.Printf("Connected to the database\n")
	}

	//FRONTEND_LOGIN is the username to be created by this script. Do not set if it's already been created
	frontendLogin := os.Getenv("FRONTEND_LOGIN")
	installConfig.FrontendLogin = frontendLogin
	frontendPassword := os.Getenv("FRONTEND_PASSWORD")
	installConfig.FrontendPassword = frontendPassword
	installConfig.FrontendRepeatPassword = frontendPassword

	dsPath := os.Getenv("DATASOURCE_PATH") // datasource location in your container. For default value look in source code
	if dsPath != "" {
		installConfig.DsFolder = dsPath
	}
	oidcId := os.Getenv("OPENID_CONNECT_CLIENT_ID") //openid connect client id for dex.
	if oidcId != "" {
		installConfig.ExternalDexID = oidcId
	}
	oidcSecret := os.Getenv("OPENID_CONNECT_CLIENT_SECRET") //openid connect client secret for dex.
	if oidcSecret != "" {
		installConfig.ExternalDexSecret = oidcSecret
	}
	// Log install messages to the console.
	log.Printf("installConfig:\n")
	printConfig(installConfig)

	err = lib.Install(context.Background(), (*install.InstallConfig)(installConfig), lib.INSTALL_ALL,
		func(event *lib.InstallProgressEvent) {
			log.Printf("Installing %v/%v: %v", event.Progress, 100,  event.Message)
		})
	if err != nil {
		log.Printf("Installation Error: %v", err)
	}
	log.Printf("Installing 100/100: Success!\n")

	//Creating installed flag
	emptyFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	_ = emptyFile.Close()
	execveCells()
}






	
	
	

