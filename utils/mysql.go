package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func FindMySQLInstallDir() (string, error) {
	var cmd = exec.Command("which", "mysqld")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var mysqldPath = out.String()
	if strings.Compare(mysqldPath, "") == 0 {
		return "", nil
	}

	cmd = exec.Command("dirname", mysqldPath)
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	var mysqlBinPath = out.String()
	cmd = exec.Command("dirname", mysqlBinPath)
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(out.String(), "\n"), nil
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func InitMySQLConfigFile(port int, user string, password string,
	mysqlDir string, sandbox string, filename string) error {
	format := `
[mysql]
prompt='mysql [\h] {\u} (\d) > '
#

[client]
user               = %s
password           = %s
port               = %d
socket             = /tmp/mysql_sandbox%d.sock

[mysqld]
user               = %s
port               = %d
socket             = /tmp/mysql_sandbox%d.sock
basedir            = %s
datadir            = %s/data
tmpdir             = %s/tmp
lower_case_table_names = 0
pid-file           = %s/data/mysql_sandbox%d.pid
bind-address       = 127.0.0.1
`

	context := fmt.Sprintf(format, user, password, port, port, user, port, port, mysqlDir, sandbox, sandbox, sandbox, port)
	f, err := os.Create(filename)
	defer f.Close()

	check(err)
	_, err = f.WriteString(context)
	check(err)
	err = f.Sync()
	check(err)

	return err
}

func GetMySQLVersion(mysqlDir string) (int, int, int, error) {
	var mysql_config = mysqlDir + "/bin/mysql_config"
	var options = "--version"
	var cmd = exec.Command(mysql_config, options)
	var out bytes.Buffer
	cmd.Stdout = &out
	var err = cmd.Run()
	if err != nil {
		return 0, 0, 0, err
	} else {
		var versions = strings.Split(strings.TrimSuffix(out.String(), "\n"), ".")
		major, _ := strconv.Atoi(versions[0])
		minor, _ := strconv.Atoi(versions[1])
		rev, _ := strconv.Atoi(versions[2])
		return major, minor, rev, nil
	}
}

func MySQLInstallDB(mysqlDir string, dataDir string) error {
	var mysql_install_db = mysqlDir + "/scripts/mysql_install_db"
	var option1 = "--basedir=" + mysqlDir
	var option2 = "--datadir=" + dataDir
	var cmd = exec.Command(mysql_install_db, option1, option2)
	var out bytes.Buffer
	cmd.Stdout = &out
	var err = cmd.Run()
	return err
}

func MySQLInstallReplication(mysqlDir string, installPath string, port int) {
	masterDir := installPath + "/master"
	slaveDir := installPath + "/slave"
	masterDataDir := masterDir + "/data"
	slaveDataDir := slaveDir + "/data"
	masterCnf := masterDir + "/my.sandbox.cnf"
	slaveCnf := slaveDir + "/my.sandbox.cnf"

	/** install master */
	os.MkdirAll(masterDataDir, 0777)
	err := MySQLInstallDB(mysqlDir, masterDataDir)
	check(err)
	err = InitMySQLConfigFile(port, "dbscale", "dbscale", mysqlDir, masterDir, masterCnf)
	check(err)

	/** install slave */
	os.MkdirAll(slaveDataDir, 0777)
	err = MySQLInstallDB(mysqlDir, slaveDataDir)
	check(err)
	err = InitMySQLConfigFile(port+1, "dbscale", "dbscale", mysqlDir, slaveDir, slaveCnf)
	check(err)
}
