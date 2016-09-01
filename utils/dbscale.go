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
partition = partition2`
	config = fmt.Sprintf(config, "log/dbscale.log", dbUser, dbPassword, dbscalePort, dbUser, dbPassword, mysqlStartPort, dbUser, dbPassword, mysqlStartPort+1, dbUser, dbPassword, mysqlStartPort+2, dbUser, dbPassword, mysqlStartPort+3, dbUser, dbPassword, dbUser, dbPassword, mysqlStartPort+4, dbUser, dbPassword, mysqlStartPort+5, dbUser, dbPassword, dbUser, dbPassword)
	return config
}

func InstallDBScaleConfig(dbUser string, dbPassword string, installPath string, mysqlStartPort int, dbscalePort int) {
	config := InitDBScaleConfig(dbUser, dbPassword, installPath, mysqlStartPort, dbscalePort)
	configPath := installPath + "/dbscale/dbscale.conf"
	err := ioutil.WriteFile(configPath, []byte(config), 0644)
	Check(err)
}

func StartDBScale(installPath string) {
	cmd := exec.Command(installPath+"/dbscale/dbscale-service.sh", "start")
	cmd.Dir = installPath + "/dbscale"
	err := cmd.Run()
	Check(err)
}

func InitStopAndStartDBScaleScripts(installPath string, startAndStopScript map[string]string) {
	startAndStopScript["startScript"] = installPath + "/dbscale/dbscale-service.sh start"
	startAndStopScript["stopScript"] = installPath + "/dbscale/dbscale-service.sh stop"
}

func InstallStartAndStopDBscaleScripts(installPath string) {
	/*** init stop&start scripts ***/
	startAndStopScript := make(map[string]string)
	InitStopAndStartDBScaleScripts(installPath, startAndStopScript)

	/*** install start scripts ***/
	startScriptPath := installPath + "/dbscale-start.sh"
	startf, err := os.Create(startScriptPath)
	Check(err)
	startf.Write([]byte(startAndStopScript["startScript"]))
	err = startf.Chmod(0755)
	Check(err)
	startf.Close()

	/*** install stop scripts ***/
	stopScriptPath := installPath + "/dbscale-stop.sh"
	stopf, err := os.Create(stopScriptPath)
	Check(err)
	stopf.Write([]byte(startAndStopScript["stopScript"]))
	err = stopf.Chmod(0755)
	Check(err)
	stopf.Close()
}
