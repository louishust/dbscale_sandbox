package utils

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

func InitGrantScripts(scripts map[string]string, options map[string]string) {
	dbUser := options["dbUser"]
	rwUser := options["rwUser"]
	roUser := options["roUser"]
	remoteAccess := options["remoteAccess"]
	dbPassword := options["dbPassword"]
	replUser := options["replUser"]
	replPassword := options["replPassword"]
	grantsMySQLFormat := `use mysql;
set password=password(%s);
grant all on *.* to %s@'%s' identified by '%s';
grant all on *.* to %s@'localhost' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to %s@'localhost' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to %s@'%s' identified by '%s';
grant SELECT,EXECUTE on *.* to %s@'%s' identified by '%s';
grant SELECT,EXECUTE on *.* to %s@'localhost' identified by '%s';
grant REPLICATION SLAVE on *.* to %s@'%s' identified by '%s';
delete from user where password='';
delete from db where user='';
flush privileges;
create database if not exists test;
`
	grantsMysql := fmt.Sprintf(grantsMySQLFormat, dbPassword, dbUser, remoteAccess, dbPassword, dbUser, dbPassword, rwUser, dbPassword, rwUser, remoteAccess, dbPassword, roUser, remoteAccess, dbPassword, roUser, dbPassword, replUser, remoteAccess, replPassword)
	scripts["grants.mysql"] = grantsMysql
	grants576MySQLFormat := `use mysql;
set password='%s';
-- delete from tables_priv;
-- delete from columns_priv;
-- delete from db;
delete from user where user not in ('root', 'mysql.sys', 'mysqlxsys');

flush privileges;

create user %s@'%s' identified by '%s';
grant all on *.* to %s@'%s' ;

create user %s@'localhost' identified by '%s';
grant all on *.* to %s@'localhost';

create user %s@'localhost' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
     SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
     on *.* to %s@'localhost';

create user %s@'%s' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to %s@'%s';
     
create user %s@'%s' identified by '%s';
create user %s@'localhost' identified by '%s';
create user %s@'%s' identified by '%s';
grant SELECT,EXECUTE on *.* to %s@'%s';
grant SELECT,EXECUTE on *.* to %s@'localhost';
grant REPLICATION SLAVE on *.* to %s@'%s';
create schema if not exists test;
`
	grants576Mysql := fmt.Sprintf(grants576MySQLFormat, dbPassword, dbUser, remoteAccess, dbPassword, dbUser, remoteAccess, dbUser, dbPassword, dbUser, rwUser, dbPassword, rwUser, rwUser, remoteAccess, dbPassword, rwUser, remoteAccess, roUser, remoteAccess, dbPassword, roUser, dbPassword, replUser, remoteAccess, replPassword, roUser, remoteAccess, roUser, replUser, remoteAccess)
	scripts["grants_5_7_6.mysql"] = grants576Mysql
}

func MySQLInstallGrantFile(installPath string, code string) {
	grantsFilePath := installPath + "/mysql.grants"

	grantsCode := []byte(code)

	grantsFile, err := os.Create(grantsFilePath)
	check(err)

	_, err = grantsFile.Write(grantsCode)
	check(err)

	grantsFile.Close()
}

func MySQLInstallRepGrantFile(grantsFilePath string, dsn string) {
	grantsInstallSQL := "source " + grantsFilePath
	RunOperat(dsn, grantsInstallSQL)
}

func InitGrantOptions(options map[string]string) {
	options["dbUser"] = "dbscale"
	options["dbPassword"] = "dbscale"
	options["remoteAccess"] = "127.%"
	options["roUser"] = "dbscale_ro"
	options["rwUser"] = "dbscale_rw"
	options["replUser"] = "rdbscale"
	options["replPassword"] = "rdbscale"
}

func RunOperat(dsn string, stmt string) {
	db, err := sql.Open("mysql", "dsn")
	check(err)
	_, exec_err := db.Exec(stmt)
	check(exec_err)
	db.Close()
}
