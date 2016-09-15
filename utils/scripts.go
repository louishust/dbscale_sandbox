package utils

var config = `[mysql]
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
bind-address       = 0.0.0.0
gtid_mode          = on
enforce-gtid-consistency = 1
net_write_timeout=1800
net_read_timeout=1800
max_allowed_packet=16777216
skip_name_resolve=1
log-bin=bin
log-slave-updates
server-id          = %d
skip_slave_start
innodb_buffer_pool_size = 5242880
innodb_read_io_threads=1
innodb_write_io_threads=1
innodb_purge_threads=1
performance_schema=OFF
`

var startScript = `#!/bin/bash
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

var stopScript = `#!/bin/bash
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

var sendKillScript = `#!/bin/bash
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

var useScript = `
#!/bin/bash
export LD_LIBRARY_PATH=%s/lib:%s/lib/mysql:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=%s/lib:%s/lib/mysql:$DYLD_LIBRARY_PATH
SBDIR="%s"
BASEDIR=%s
[ -z "$MYSQL_EDITOR" ] && MYSQL_EDITOR="$BASEDIR/bin/mysql"
HISTDIR=
[ -z "$HISTDIR" ] && HISTDIR=$SBDIR
export MYSQL_HISTFILE="$HISTDIR/.mysql_history"
PIDFILE="$SBDIR/data/mysql_sandbox%d.pid"
if [ -f "$SBINSTR" ]
then
    echo "[%sbasename $0%s] - %sdate "%s"%s - $@" >> $SBINSTR
fi

if [ -f $PIDFILE ]
then
    $MYSQL_EDITOR --defaults-file=$SBDIR/my.sandbox.cnf $MYCLIENT_OPTIONS "$@"
fi
`
