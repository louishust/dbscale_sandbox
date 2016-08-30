package main

import (
	"flag"
	"fmt"
	utils "github.com/louishust/dbscale-sandbox/utils"
	"os"
	"os/user"
)

var (
	installPath        *string
	mysqlDirPath       *string
	mysqlStartPort     *int
	dbscalePackagePath *string
	dbscalePort        *int
)

func init_options() {
	var curUser, _ = user.Current()
	var defaultInstallPath = curUser.HomeDir + "/sandboxes"
	var defaultMySQLDirPath, _ = utils.FindMySQLInstallDir()
	var defaultStartPort = 3210
	var defaultDBSalePort = 13001
	installPath = flag.String("install-path", defaultInstallPath, "path to install")
	mysqlDirPath = flag.String("mysql-dir", defaultMySQLDirPath, "mysql installed directory")
	mysqlStartPort = flag.Int("mysql-start-port", defaultStartPort, "mysql start port")
	dbscalePackagePath = flag.String("dbscale-package-path", "", "DBScale package path")
	dbscalePort = flag.Int("dbscale-port", defaultDBSalePort, "DBScale port")
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	init_options()
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *dbscalePackagePath == "" {
		fmt.Println("dbscale-package-path must be declare!")
		os.Exit(1)
	}

	rep1Dir := *installPath + "/rep_mysql_sandbox1"
	rep2Dir := *installPath + "/rep_mysql_sandbox2"
	authDir := *installPath + "/auth_mysql_sandbox"

	instanceDir2Port := map[string]int{
		authDir + "/master": *mysqlStartPort,
		authDir + "/slave":  *mysqlStartPort + 1,
		rep1Dir + "/master": *mysqlStartPort + 2,
		rep1Dir + "/slave":  *mysqlStartPort + 3,
		rep2Dir + "/master": *mysqlStartPort + 4,
		rep2Dir + "/slave":  *mysqlStartPort + 5}

	ret1, _ := exists(rep1Dir)
	ret2, _ := exists(rep2Dir)
	ret3, _ := exists(authDir)

	if ret1 || ret2 || ret3 {
		fmt.Println(rep1Dir + " or " + rep2Dir + " or " + authDir + " already exists!!!")
		os.Exit(1)
	}

	os.MkdirAll(rep1Dir, 0777)
	os.MkdirAll(rep2Dir, 0777)
	os.MkdirAll(authDir, 0777)

	utils.MySQLInstallReplication(*mysqlDirPath, instanceDir2Port)

	utils.InstallMySQLStartScripts(*mysqlDirPath, *installPath, instanceDir2Port)
	utils.StartMySQL(*installPath)

	/** init grants options **/
	options := make(map[string]string)
	utils.InitGrantOptions(options)

	grantsCode := utils.MySQLInstallGrantFile(*mysqlDirPath, *installPath, options)
	utils.MySQLInstallRepGrantFile(grantsCode, instanceDir2Port)

	utils.InstallDBScale(*dbscalePackagePath, *installPath)
	utils.InstallDBScaleConfig(options["dbUser"], options["dbPassword"], *installPath, *mysqlStartPort, *dbscalePort)
	utils.StartDBScale(*installPath)
}
