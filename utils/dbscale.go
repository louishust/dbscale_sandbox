package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func InitDBScaleServiceScript(installPath string) {
	serviceScript := `#!/bin/bash
PROG=dbscale
DBSCALE_PATH=%s/dbscale # Need to modify
DBSCALE_PID_FILE=$DBSCALE_PATH/dbscale.pid
DBSCALE_PID=%scat $DBSCALE_PID_FILE 2>/dev/null%s
START_TIMEOUT=60
STOP_TIMEOUT1=10
STOP_TIMEOUT2=5

force=0

check_pid() {
    [ -z $1 ] && return 1
    [ -d "/proc/$1" ] && return 0 || return 1
}

dbscale_start() {
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$DBSCALE_PATH/libs
    cd $DBSCALE_PATH

    ./$PROG &
    DBSCALE_PID=$!
    sleep 1

    start_time=%sdate '+%s'%s
    while true; do
        current_time=%sdate '+%s'%s
        if (($current_time - $start_time > $START_TIMEOUT)); then # start time out
            return
        fi
        if [ ! -f "dbscale.pid" ]; then
            continue
        fi
        if [ "X$DBSCALE_PID" == "X"%scat dbscale.pid%s ]; then
            return
        fi
    done
}

dbscale_stop() {
    kill -TERM $DBSCALE_PID >/dev/null 2>&1
    start_time=%sdate '+%s'%s
    while true; do
        current_time=%sdate '+%s'%s
        if (($current_time - $start_time > $STOP_TIMEOUT1)); then
            break
        fi
        if ! check_pid $DBSCALE_PID; then
            return
        fi
    done

    kill -KILL $DBSCALE_PID >/dev/null 2>&1
    sleep $STOP_TIMEOUT2
}

if [ ! -d "$DBSCALE_PATH" -o ! -f "$DBSCALE_PATH/$PROG" ]; then
    echo "The path of dbscale is not correct."
    exit 1
fi

case "$1" in
    start)
        if ! check_pid $DBSCALE_PID; then
            echo -e "Starting DBScale...\c"
            dbscale_start
            check_pid $DBSCALE_PID && echo "done." || echo "fail."
        else
            echo "DBScale has already been running."
        fi
        ;;
    stop)
        if check_pid $DBSCALE_PID; then
            echo -e "Stopping DBScale...\c"
            dbscale_stop
            check_pid $DBSCALE_PID && echo "fail." || echo "done."
        else
            echo "DBScale is not running."
        fi
        ;;
    status)
        if check_pid $DBSCALE_PID; then
            echo "DBScale is running."
            exit 0
        else
            echo "DBScale is not running."
            exit 1
        fi
        ;;
    *)
        echo "Usage: $0 {start|stop|status}"
        exit 1
esac
exit 0
`
	backQuotes := "`"
	serviceScript = fmt.Sprintf(serviceScript, installPath, backQuotes, backQuotes, backQuotes, "%s", backQuotes, backQuotes, "%s", backQuotes, backQuotes, backQuotes, backQuotes, "%s", backQuotes, backQuotes, "%s", backQuotes)
	serviceScriptPath := installPath + "/dbscale/dbscale-service.sh"
	err := os.Remove(serviceScriptPath)
	Check(err)
	f, err := os.Create(serviceScriptPath)
	Check(err)
	f.Write([]byte(serviceScript))
	err = f.Chmod(0755)
	Check(err)
	f.Close()
}

func InstallDBScale(dbscalePackagePath string, installPath string) {
	fmt.Println("Installing DBScale...")
	Decompress(dbscalePackagePath, installPath)
	InitDBScaleServiceScript(installPath)
}

func InitDBScaleConfig(dbUser string, dbPassword string, installPath string, mysqlStartPort int, dbscalePort int) string {
	config := `[main]
driver = mysql
log-level = INFO
backlog = 10240
log-file = %s
real-time-queries = 2
admin-user = %s
admin-password = %s
max-replication-delay = 500
default-session-variables = CHARACTER_SET_CLIENT:CHARACTER_SET_RESULTS:CHARACTER_SET_CONNECTION:NET_READ_TIMEOUT:TIME_ZONE:SQL_SAFE_UPDATES:SQL_MODE:AUTOCOMMIT:TX_ISOLATION:SQL_SELECT_LIMIT
support-gtid=1
authenticate-source = auth
is-auth-schema = 1
use-partial-parse = 1
lower-case-table-names = 0
thread-pool-min = 50
thread-pool-max = 80
thread-pool-low = 40
backend-thread-pool-max= 80
handler-thread-pool-max=80
max-fetchnode-ready-rows-size=1000000
auto-inc-lock-mode=0
enable-get-rep-connection = 1
enable-session-swap=0

[driver mysql]
type = MySQLDriver
port = %d
bind-address = 0.0.0.0

[catalog def]
data-source = ds_catalog

[data-source ds_catalog]
type = replication
master = p1m-4-50-18-40
slave = p1s-4-50-18-40
user = %s
password = %s
load-balance-strategy = MASTER-SLAVES

[data-server auth_m]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-server auth_s]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-source auth]
type = replication
master = auth_m-4-50-18-40
slave  = auth_s-4-50-18-40
load-balance-strategy = MASTER-SLAVES

[data-server p1m]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-server p1s]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-source partition1]
type = replication
master = p1m-10-50-20-40
slave = p1s-10-50-20-40
user = %s
password = %s
load-balance-strategy = MASTER-SLAVES

[data-server p2m]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-server p2s]
host = 127.0.0.1
port = %d
user = %s
password = %s

[data-source partition2]
type = replication
master = p2m-10-50-20-40
slave = p2s-10-50-20-40
user = %s
password = %s
load-balance-strategy = MASTER-SLAVES

[partition-scheme test]
type=hash
virtual-weight = 1:1
partition = partition1
partition = partition2

[table test.part]
type = partition
pattern=.*
partition-scheme = test
partition-key = id 
`

	config = fmt.Sprintf(config, "log/dbscale.log", dbUser, dbPassword, dbscalePort, dbUser, dbPassword, mysqlStartPort, dbUser, dbPassword, mysqlStartPort+1, dbUser, dbPassword, mysqlStartPort+2, dbUser, dbPassword, mysqlStartPort+3, dbUser, dbPassword, dbUser, dbPassword, mysqlStartPort+4, dbUser, dbPassword, mysqlStartPort+5, dbUser, dbPassword, dbUser, dbPassword)
	return config
}

func InstallDBScaleConfig(dbUser string, dbPassword string, installPath string, mysqlStartPort int, dbscalePort int) {
	fmt.Println("Initing DBScale Configure...")
	config := InitDBScaleConfig(dbUser, dbPassword, installPath, mysqlStartPort, dbscalePort)
	configPath := installPath + "/dbscale/dbscale.conf"
	err := ioutil.WriteFile(configPath, []byte(config), 0644)
	Check(err)
}

func StartDBScale(installPath string) {
	fmt.Println("Starting DBScale...")
	cmd := exec.Command(installPath+"/dbscale/dbscale-service.sh", "start")
	cmd.Dir = installPath + "/dbscale"
	err := cmd.Run()
	Check(err)
}

func InitDBScaleScripts(installPath string, mysqlDirPath string, dbUser string, dbPassword string, dbscalePort int, DBScaleScript map[string]string) {
	DBScaleScript["dbscale-start.sh"] = installPath + "/dbscale/dbscale-service.sh start"
	DBScaleScript["dbscale-stop.sh"] = installPath + "/dbscale/dbscale-service.sh stop"
	loginScript := "%s/bin/mysql -u%s -p%s -h127.0.0.1 -P%d"
	loginScript = fmt.Sprintf(loginScript, mysqlDirPath, dbUser, dbPassword, dbscalePort)
	DBScaleScript["loginDBScale"] = loginScript
}

func InstallDBscaleScripts(installPath string, mysqlDirPath string, dbUser string, dbPassword string, dbscalePort int) {
	fmt.Println("Installing DBScale Scripts And Sandbox Scripts.")
	/*** init stop&start scripts ***/
	DBScaleScript := make(map[string]string)
	InitDBScaleScripts(installPath, mysqlDirPath, dbUser, dbPassword, dbscalePort, DBScaleScript)

	/*** install DBScale scripts ***/
	for scriptName, script := range DBScaleScript {
		scriptPath := installPath + "/" + scriptName
		scriptf, err := os.Create(scriptPath)
		Check(err)
		scriptf.Write([]byte(script))
		err = scriptf.Chmod(0755)
		Check(err)
		scriptf.Close()
	}
}

func InitPartitionData(dbscalePort int, dbUser string, dbPassword string) {
	fmt.Println("Initing Partition Data.")
	stmts := []string{
		"create table part_tb01 (id int primary key, c1 int, c2 varchar(20)) engine=innodb",
		"create table part_tb02 (id int primary key, c1 int, c2 varchar(20)) engine=innodb",
		"insert into part_tb01 values (1, 1, 'hello world.')",
		"insert into part_tb01 values (2, 2, 'welecome to dbscale.')",
		"insert into part_tb01 values (3, 3, 'this is a demo partition table.')",
		"insert into part_tb01 values (4, 4, 'plz try and have fun.')",
	}
	for i := 5; i < 101; i++ {
		sql := "insert into part_tb01 values (%d, %d, 'this is test text%d')"
		sql = fmt.Sprintf(sql, i, i, i)
		stmts = append(stmts, sql)
	}
	dsn := "%s:%s@tcp(127.0.0.1:%d)/test"
	dsn = fmt.Sprintf(dsn, dbUser, dbPassword, dbscalePort)
	RunOperat(dsn, stmts)
}

func InstallAndStartScale(mysqlDirPath string, dbscalePackagePath string, installPath string, mysqlStartPort int, dbscalePort int) {
	InstallDBScale(dbscalePackagePath, installPath)
	InstallDBScaleConfig(Options["dbUser"], Options["dbPassword"], installPath, mysqlStartPort, dbscalePort)
	StartDBScale(installPath)
	InstallDBscaleScripts(installPath, mysqlDirPath, Options["dbUser"], Options["dbPassword"], dbscalePort)

	InstallScripts4All(installPath)

	InitPartitionData(dbscalePort, Options["dbUser"], Options["dbPassword"])
}
