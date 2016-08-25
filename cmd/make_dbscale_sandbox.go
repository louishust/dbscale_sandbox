package main

import (
	"flag"
	"fmt"
	utils "github.com/louishust/dbscale-sandbox/utils"
	"os"
	"os/user"
)

var (
	installPath  *string
	mysqlDirPath *string
	start_port   = 3210
)

func init_options() {
	var curUser, _ = user.Current()
	var defaultInstallPath = curUser.HomeDir + "/sandboxes"
	var defaultMySQLDirPath, _ = utils.FindMySQLInstallDir()
	installPath = flag.String("install-path", defaultInstallPath, "path to install")
	mysqlDirPath = flag.String("mysql-dir", defaultMySQLDirPath, "mysql installed directory")
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

	rep1Dir := *installPath + "/rep_mysql_sandbox1"
	rep2Dir := *installPath + "/rep_mysql_sandbox2"
	authDir := *installPath + "/auth_mysql_sandbox"

	instanceDir2Port := map[string]int{
		rep1Dir + "/master": start_port,
		rep1Dir + "/slave":  start_port + 1,
		rep2Dir + "/master": start_port + 2,
		rep2Dir + "/slave":  start_port + 3,
		authDir + "/master": start_port + 4,
		authDir + "/slave":  start_port + 5}

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

	grantsCode := utils.MySQLInstallGrantFile(*mysqlDirPath, *installPath)
	utils.MySQLInstallRepGrantFile(grantsCode, instanceDir2Port)
}
