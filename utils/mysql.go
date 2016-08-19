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

func Check(e error) {
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

	Check(err)
	_, err = f.WriteString(context)
	Check(err)
	err = f.Sync()
	Check(err)

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

func MySQLInstallReplication(mysqlDir string, partation1InstallPath string, partation2InstallPath string, port int) {
	/** partation1 MySQL path init **/
	partation1MasterDir := partation1InstallPath + "/master"
	partation1SlaveDir := partation1InstallPath + "/slave"

	partation1MasterDataDir := partation1MasterDir + "/data"
	partation1SlaveDataDir := partation1SlaveDir + "/data"

	partation1MasterCnf := partation1MasterDir + "/my.sandbox.cnf"
	partation1SlaveCnf := partation1SlaveDir + "/my.sandbox.cnf"

	/** partation2 MySQL path init **/
	partation2MasterDir := partation2InstallPath + "/master"
	partation2SlaveDir := partation2InstallPath + "/slave"

	partation2MasterDataDir := partation2MasterDir + "/data"
	partation2SlaveDataDir := partation2SlaveDir + "/data"

	partation2MasterCnf := partation2MasterDir + "/my.sandbox.cnf"
	partation2SlaveCnf := partation2SlaveDir + "/my.sandbox.cnf"

	/** install partation1 master */
	os.MkdirAll(partation1MasterDataDir, 0777)
	err := MySQLInstallDB(mysqlDir, partation1MasterDataDir)
	Check(err)
	err = InitMySQLConfigFile(port, "dbscale", "dbscale", mysqlDir, partation1MasterDir, partation1MasterCnf)
	Check(err)

	/** install partation1 slave */
	os.MkdirAll(partation1SlaveDataDir, 0777)
	err = MySQLInstallDB(mysqlDir, partation1SlaveDataDir)
	Check(err)
	err = InitMySQLConfigFile(port+1, "dbscale", "dbscale", mysqlDir, partation1SlaveDir, partation1SlaveCnf)
	Check(err)

	/** install partation2 master */
	os.MkdirAll(partation2MasterDataDir, 0777)
	err = MySQLInstallDB(mysqlDir, partation2MasterDataDir)
	Check(err)
	err = InitMySQLConfigFile(port, "dbscale", "dbscale", mysqlDir, partation2MasterDir, partation2MasterCnf)
	Check(err)

	/** install partation2 slave */
	os.MkdirAll(partation2SlaveDataDir, 0777)
	err = MySQLInstallDB(mysqlDir, partation2SlaveDataDir)
	Check(err)
	err = InitMySQLConfigFile(port+1, "dbscale", "dbscale", mysqlDir, partation2SlaveDir, partation2SlaveCnf)
	Check(err)
}

func InitGrantScripts(scripts map[string]string) {
	/** init grants options **/
	options := make(map[string]string)
	InitGrantOptions(options)

	/** get grants options **/
	dbUser := options["dbUser"]
	rwUser := options["rwUser"]
	roUser := options["roUser"]
	remoteAccess := options["remoteAccess"]
	dbPassword := options["dbPassword"]
	replUser := options["replUser"]
	replPassword := options["replPassword"]

	/** init grants code **/
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

	/** init grants576 code **/
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

func MySQLInstallGrantFile(mysqlDirPath string, installPath string) string {
	/** init grants scripts **/
	scripts := make(map[string]string)
	InitGrantScripts(scripts)

	/** judge version **/
	verP1, verP2, verP3, err := GetMySQLVersion(mysqlDirPath)
	Check(err)

	var code string

	if (verP1*256*256 + verP2*256 + verP3) >= (5*256*256 + 7*256 + 6) {
		code = scripts["grants_5_7_6.mysql"]
	} else {
		code = scripts["grants.mysql"]
	}

	grantsFilePath := installPath + "/mysql.grants"

	grantsCode := []byte(code)

	grantsFile, err := os.Create(grantsFilePath)
	Check(err)

	_, err = grantsFile.Write(grantsCode)
	Check(err)

	grantsFile.Close()
	return grantsFilePath
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
	Check(err)
	_, exec_err := db.Exec(stmt)
	Check(exec_err)
	db.Close()
}
