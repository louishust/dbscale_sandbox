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
	var mysqlDir = "/home/vagrant/sandbox/5.6.26"
	t.Log(mysqlDir)
	retChan := make(chan error, 1)
	go MySQLInstallDB(mysqlDir, "/tmp/data", retChan)
	err := <-retChan
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
	retChan := make(chan error, 1)
	go InitMySQLConfigFile(3306, "msandbox", "msandbox", "/home/vagrant/sandbox/5.6.26", "/tmp/msandbox_1234", "my.sandbox.cnf", retChan)

	err := <-retChan
	if err != nil {
		t.Error(err)
	}
}
