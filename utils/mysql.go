package utils

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"os/exec"
	"path"
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
	mysqlDir string, sandbox string, filename string, retChan chan error) {
	format := `[mysql]
prompt='mysql [\h] {\u} (\d) > '
#

[client]
user               = %s
password           = %s
port               = %d
socket             = %s/mysql_sandbox%d.sock

[mysqld]
user               = %s
port               = %d
socket             = %s/mysql_sandbox%d.sock
basedir            = %s
datadir            = %s/data
tmpdir             = %s/tmp
lower_case_table_names = 1
pid-file           = %s/data/mysql_sandbox%d.pid
bind-address       = 127.0.0.1
gtid_mode          = on
enforce-gtid-consistency = 1
net_write_timeout=1800
net_read_timeout=1800
max_allowed_packet=16777216
skip_name_resolve=1
log-bin=bin
log-slave-updates
server-id          = %d
`

	context := fmt.Sprintf(format, user, password, port, sandbox, port, user, port, sandbox, port, mysqlDir, sandbox, sandbox, sandbox, port, port)
	f, err := os.Create(filename)
	defer f.Close()

	Check(err)
	_, err = f.WriteString(context)
	Check(err)
	err = f.Sync()
	Check(err)

	retChan <- err
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

func MySQLInstallDB(mysqlDir string, dataDir string, retChan chan error) {
	var mysql_install_db = mysqlDir + "/scripts/mysql_install_db"
	var option1 = "--basedir=" + mysqlDir
	var option2 = "--datadir=" + dataDir
	var cmd = exec.Command(mysql_install_db, option1, option2)
	var out bytes.Buffer
	cmd.Stdout = &out
	var err = cmd.Run()
	retChan <- err
}

func MySQLInstallReplication(mysqlDir string, installPath string, mysqlPackagePath string, instanceDir2Port map[string]int) {
	/*** judge weather need to decompress MySQL ***/
	if mysqlPackagePath != "" {
		Decompress(mysqlPackagePath, installPath)
		mysqlDecompressPath := strings.Split(path.Base(mysqlPackagePath), ".tar")[0]
		cmd := exec.Command("ln", "-s", mysqlDecompressPath, installPath+"/mysql")
		err := cmd.Run()
		Check(err)
	}

	retChan := make(chan error, 12)

	/*** Install MySQL and config ***/
	for dir, port := range instanceDir2Port {
		dataDir := dir + "/data"
		tmpDir := dir + "/tmp"
		cnfPath := dir + "/my.sandbox.cnf"
		os.MkdirAll(dataDir, 0777)
		os.MkdirAll(tmpDir, 0777)
		go MySQLInstallDB(mysqlDir, dataDir, retChan)
		go InitMySQLConfigFile(port, "dbscale", "dbscale", mysqlDir, dir, cnfPath, retChan)
	}

	/** check return channel **/
	for i := 0; i < 12; i++ {
		err := <-retChan
		Check(err)
	}
}

func InitMySQLScripts(mysqlDirPath string, instanceDir string, port int, scriptsDict map[string]string) {
	startScript := `#!/bin/bash
BASEDIR='%s'
export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=$BASEDIR_/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
MYSQLD_SAFE="$BASEDIR/bin/mysqld_safe"
SBDIR="%s"
PIDFILE="$SBDIR/data/mysql_sandbox%d.pid"

if [ ! -f $MYSQLD_SAFE ]
then
    echo "mysqld_safe not found in $BASEDIR/bin/"
    exit 1
fi
MYSQLD_SAFE_OK=%ssh -n $MYSQLD_SAFE 2>&1%s
if [ "$MYSQLD_SAFE_OK" != "" ]
then
    echo "$MYSQLD_SAFE has errors"
    echo "((( $MYSQLD_SAFE_OK )))"
    exit 1
fi

is_running()
{
    if [ -f $PIDFILE ]
    then
        MYPID=$(cat $PIDFILE)
        ps -p $MYPID | grep $MYPID
    fi
}

TIMEOUT=180
if [ -n "$(is_running)" ]
then
    echo "sandbox server already started (found pid file $PIDFILE)"
else
    if [ -f $PIDFILE ]
    then
        # Server is not running. Removing stale pid-file
        rm -f $PIDFILE
    fi
    CURDIR=%spwd%s
    cd $BASEDIR
    $MYSQLD_SAFE --defaults-file=$SBDIR/my.sandbox.cnf $@ > /dev/null 2>&1 &
    cd $CURDIR
    ATTEMPTS=1
    while [ ! -f $PIDFILE ] 
    do
        ATTEMPTS=$(( $ATTEMPTS + 1 ))
        echo -n "."
        if [ $ATTEMPTS = $TIMEOUT ]
        then
            break
        fi
        sleep 1
    done
fi

if [ -f $PIDFILE ]
then
    echo " sandbox server started"
    #if [ -f $SBDIR/needs_reload ]
    #then
    #    if [ -f $SBDIR/rescue_mysql_dump.sql ]
    #    then
    #        $SBDIR/use mysql < $SBDIR/rescue_mysql_dump.sql
    #    fi
    #    rm $SBDIR/needs_reload
    #fi
else
    echo " sandbox server not started yet"
    exit 1
fi
`
	backQuotes := "`"
	scriptsDict["start"] = fmt.Sprintf(startScript, mysqlDirPath, instanceDir, port, backQuotes, backQuotes, backQuotes, backQuotes)
	stopScript := `#!/bin/bash
BASEDIR="%s"
SBDIR="%s"
export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
MYSQL_ADMIN="$BASEDIR/bin/mysqladmin"
PIDFILE="$SBDIR/data/mysql_sandbox%d.pid"

is_running()
{
    if [ -f $PIDFILE ]
    then
        MYPID=$(cat $PIDFILE)
        ps -p $MYPID | grep $MYPID
    fi
}

if [ -n "$(is_running)" ]
then
    $MYSQL_ADMIN --defaults-file=$SBDIR/my.sandbox.cnf $MYCLIENT_OPTIONS shutdown
    sleep 1
else
    if [ -f $PIDFILE ]
    then
        rm -f $PIDFILE
    fi
fi

if [ -n "$(is_running)" ]
then
    # use the send_kill script if the server is not responsive
    $SBDIR/send_kill
fi
`
	scriptsDict["stop"] = fmt.Sprintf(stopScript, mysqlDirPath, instanceDir, port)
	sendKillScript := `#!/bin/bash
SBDIR="%s"
PIDFILE="$SBDIR/data/mysql_sandbox%d.pid"
TIMEOUT=30

is_running()
{
    if [ -f $PIDFILE ]
    then
        MYPID=$(cat $PIDFILE)
        ps -p $MYPID | grep $MYPID
    fi
}


if [ -n "$(is_running)" ]
then
    MYPID=%scat $PIDFILE%s
    echo "Attempting normal termination --- kill -15 $MYPID"
    kill -15 $MYPID
    # give it a chance to exit peacefully
    ATTEMPTS=1
    while [ -f $PIDFILE ]
    do
        ATTEMPTS=$(( $ATTEMPTS + 1 ))
        if [ $ATTEMPTS = $TIMEOUT ]
        then
            break
        fi
        sleep 1
    done
    if [ -f $PIDFILE ]
    then
        echo "SERVER UNRESPONSIVE --- kill -9 $MYPID"
        kill -9 $MYPID
        rm -f $PIDFILE
    fi
else
    # server not running - removing stale pid-file
    if [ -f $PIDFILE ]
    then
        rm -f $PIDFILE
    fi
fi
`
	scriptsDict["send_kill"] = fmt.Sprintf(sendKillScript, instanceDir, port, backQuotes, backQuotes)
}

func InitMySQLScript4All(installPath string, scriptsDict map[string]string) {
	startAllMySQLScript := `#!/bin/bash
instanceScript=%sfind %s -maxdepth 3 -mindepth 2 -type f -name start%s
for i in $instanceScript;
do
    $i $@
done
`
	backQuotes := "`"
	scriptsDict["startallmysql"] = fmt.Sprintf(startAllMySQLScript, backQuotes, installPath, backQuotes)
	stopAllMySQLScript := `#!/bin/bash
instanceScript=%sfind %s -maxdepth 3 -mindepth 2 -type f -name stop%s
for i in $instanceScript;
do
    $i $@
done
`
	scriptsDict["stopallmysql"] = fmt.Sprintf(stopAllMySQLScript, backQuotes, installPath, backQuotes)
}

func InstallMySQLScripts(mysqlDirPath string, installPath string, instanceDir2Port map[string]int) {
	mysqlScriptsDict := make(map[string]string)
	mysqlScript4AllDict := make(map[string]string)

	/*** Install startscripts ***/
	for instanceDir, port := range instanceDir2Port {
		InitMySQLScripts(mysqlDirPath, instanceDir, port, mysqlScriptsDict)
		for scriptName, script := range mysqlScriptsDict {
			scriptFilePath := instanceDir + "/" + scriptName
			scriptFile, err := os.Create(scriptFilePath)
			Check(err)
			_, err = scriptFile.Write([]byte(script))
			Check(err)
			scriptFile.Chmod(0744)
			scriptFile.Close()
		}
	}

	/*** Install scripts4all ***/
	InitMySQLScript4All(installPath, mysqlScript4AllDict)
	for scriptName, script := range mysqlScript4AllDict {
		scriptFilePath := installPath + "/" + scriptName
		scriptFile, err := os.Create(scriptFilePath)
		Check(err)
		_, err = scriptFile.Write([]byte(script))
		Check(err)
		scriptFile.Chmod(0744)
		scriptFile.Close()
	}
}

func StartMySQL(installPath string) {
	cmd := exec.Command(installPath + "/startallmysql")
	cmd.Dir = installPath
	err := cmd.Run()
	Check(err)
}

func InitGrantScripts(scripts map[string]string, options map[string]string) {
	/** get grants options **/
	dbUser := options["dbUser"]
	rwUser := options["rwUser"]
	roUser := options["roUser"]
	remoteAccess := options["remoteAccess"]
	dbPassword := options["dbPassword"]
	replUser := options["replUser"]
	replPassword := options["replPassword"]

	/** init grants code **/
	grantsMySQLFormat := `set password=password('%s');
grant all on *.* to %s@'%s' identified by '%s' with grant option;
grant all on *.* to %s@'localhost' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER, SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE on *.* to %s@'localhost' identified by '%s';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER, SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE on *.* to %s@'%s' identified by '%s';
grant SELECT,EXECUTE on *.* to %s@'%s' identified by '%s';
grant SELECT,EXECUTE on *.* to %s@'localhost' identified by '%s';
grant REPLICATION SLAVE on *.* to %s@'%s' identified by '%s';
delete from user where password='';
delete from db where user='';
flush privileges;
create database if not exists test;`
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
grant all on *.* to %s@'%s' with grant option;

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

func MySQLInstallGrantFile(mysqlDirPath string, installPath string, options map[string]string) string {
	/*** init grants scripts ***/
	scripts := make(map[string]string)
	InitGrantScripts(scripts, options)

	/*** judge version ***/
	verP1, verP2, verP3, err := GetMySQLVersion(mysqlDirPath)
	Check(err)

	/*** install grants file ***/
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
	return code
}

func MySQLInstallRepGrantFile(grantsCode string, instanceDir2Port map[string]int) {
	for dir, port := range instanceDir2Port {
		socketPath := fmt.Sprintf("%s/mysql_sandbox%d.sock", dir, port)
		dsn := "root:@unix(" + socketPath + ")/mysql"
		stmts := strings.Split(grantsCode, "\n")
		RunOperat(dsn, stmts)
	}
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

func RunOperat(dsn string, stmts []string) {
	db, err := sql.Open("mysql", dsn)
	Check(err)
	for _, each_stmt := range stmts {
		_, err = db.Exec(each_stmt)
		Check(err)
	}
	db.Close()
}
