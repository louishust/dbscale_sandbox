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

	ret1, _ := exists(rep1Dir)
	ret2, _ := exists(rep2Dir)
	if ret1 || ret2 {
		fmt.Println(rep1Dir + " or " + rep2Dir + " already exists!!!")
		os.Exit(1)
	}

	os.MkdirAll(rep1Dir, 0777)
	os.MkdirAll(rep2Dir, 0777)

	utils.MySQLInstallReplication(*mysqlDirPath, rep1Dir, start_port)
	utils.MySQLInstallReplication(*mysqlDirPath, rep2Dir, start_port+2)
}
