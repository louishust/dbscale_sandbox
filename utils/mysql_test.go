package utils

import (
	"testing"
)

func TestFindMySQLInstallDir(t *testing.T) {
	mysqlDir, err := FindMySQLInstallDir()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(mysqlDir)
	}
}

func TestMySQLInstallDB(t *testing.T) {
	var mysqlDir, _ = FindMySQLInstallDir()
	t.Log(mysqlDir)
	var err = MySQLInstallDB(mysqlDir, "/tmp/data")
	if err != nil {
		t.Error(err)
	}
}

func TestGetMySQLVersion(t *testing.T) {
	var mysqlDir, _ = FindMySQLInstallDir()
	major, minor, rev, err := GetMySQLVersion(mysqlDir)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(major, ".", minor, ".", rev)
	}
}

func TestInitMySQLConfigFile(t *testing.T) {
	err := InitMySQLConfigFile(3306, "msandbox", "msandbox", "/usr/local/mysql", "/home/loushuai/msandbox_1234", "my.sandbox.cnf")

	if err != nil {
		t.Error(err)
	}
}
