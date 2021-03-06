# dbscale_sandbox
----------

**DBScale Sandbox** is a tool that installs one or more MySQL servers within seconds,
 easily, securely, and with full control.

Once installed, the sandbox is easily used and maintained, without using complex options.

Shard sandbox with replications can be used.


## How to get and install
---------------

```
go get github.com/louishust/dbscale_sandbox
```

## How to build
-------------

```
go build github.com/louishust/dbscale_sandbox
```


## How to make test
------------

```
go test -test.v github.com/louishust/dbscale_sandbox/utils
```

## How to use
------------

### Prepare

dbscale_sandbox must be run in "dbscale" user.

```
useradd dbscale
su - dbscale
```

### Options
```
dbscale-package-path:     DBScale package path.
dbscale-port:             DBScale port (default 13001).
install-path:             path to install (default "/userHomePath/sandboxes").
mysql-dir                 MySQL installed directory, if not declare, sandbox will auto find.
mysql-package-path        MySQL package path. If be declared, MySQL will be installed in install-path. This option priority is higher than mysql-path.
mysql-start-port          MySQL start port (default 3210). MySQL port start with mysql-start-port, and next 5 port.
```

### Example
```
./make_dbscale_sandbox -dbscale-package-path /opt/DBScale-1.5-1512.tar.gz  -mysql-package-path /opt/mysql-5.7.13-linux-glibc2.5-x86_64.tar.gz
```
