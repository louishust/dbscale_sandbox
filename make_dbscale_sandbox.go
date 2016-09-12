package main

import (
	"flag"
	"fmt"
	utils "github.com/louishust/dbscale_sandbox/utils"
	"os"
	"os/user"
)

var (
	installPath        *string
	mysqlDirPath       *string
	mysqlPackagePath   *string
	mysqlStartPort     *int
	dbscalePackagePath *string
	dbscalePort        *int
)

func init_options() {
	var curUser, _ = user.Current()
	var defaultInstallPath = curUser.HomeDir + "/sandboxes"
	var defaultStartPort = 3210
	var defaultDBSalePort = 13001
	installPath = flag.String("install-path", defaultInstallPath, "path to install")
	mysqlDirPath = flag.String("mysql-dir", "", "mysql installed directory, if not declare, sandbox will auto find")
	mysqlPackagePath = flag.String("mysql-package-path", "", "mysql package path")
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

	if *mysqlPackagePath != "" && *mysqlDirPath != "" {
		*mysqlDirPath = *installPath + "/mysql"
		fmt.Println("'mysql-package-path' higher priority than 'mysql-dir'\nnew 'mysql-dir' is %s.", *mysqlDirPath)
	} else if *mysqlPackagePath != "" {
		*mysqlDirPath = *installPath + "/mysql"
	} else {
		*mysqlDirPath, _ = utils.FindMySQLInstallDir()
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

	fmt.Println("Installing MySQL.")
	utils.MySQLInstallReplication(*mysqlDirPath, *installPath, *mysqlPackagePath, instanceDir2Port)

	fmt.Println("Installing MySQL Scripts.")
	utils.InstallMySQLScripts(*mysqlDirPath, *installPath, instanceDir2Port)

	fmt.Println("Starting MySQL.")
	utils.StartMySQL(*installPath)

	/** init grants options **/
	fmt.Println("Granting MySQL.")
	options := make(map[string]string)
	utils.InitGrantOptions(options)

	grantsCode := utils.MySQLInstallGrantFile(*mysqlDirPath, *installPath, options)
	utils.MySQLInstallRepGrantFile(grantsCode, instanceDir2Port)

	fmt.Println("Installing DBScale.")
	utils.InstallDBScale(*dbscalePackagePath, *installPath)
	fmt.Println("Initing DBScale Configure.")
	utils.InstallDBScaleConfig(options["dbUser"], options["dbPassword"], *installPath, *mysqlStartPort, *dbscalePort)
	fmt.Println("Starting DBScale.")
	utils.StartDBScale(*installPath)

	fmt.Println("Installing DBScale Scripts And Sandbox Scripts.")
	utils.InstallDBscaleScripts(*installPath, *mysqlDirPath, options["dbUser"], options["dbPassword"], *dbscalePort)

	utils.InstallScripts4All(*installPath)

	fmt.Println("Initing Partition Data.")
	utils.InitPartitionData(*dbscalePort, options["dbUser"], options["dbPassword"])

	fmt.Println("Done!")
}
