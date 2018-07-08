package main

import "C"

import (
	"github.com/liuzhishan/dman/cmd"
)

//export CApply
func CApply(dbkey, appkey, workername, info string) bool {
	return cmd.Apply(dbkey, appkey, workername, info)
}

//export CCheck
func CCheck(dbkey, appkey string) bool {
	return cmd.Check(dbkey, appkey)
}

//export CGetDbinfo
func CGetDbinfo(dbkey, appkey string) (status int, dbkey1, hostname, dbname, username, password string, port int) {
	dbkey1 = dbkey
	err, dbinfo := cmd.GetDbinfo(dbkey, appkey)
	if err != nil {
		return status, dbkey1, hostname, dbname, username, password, port
	}

	return 1, dbinfo.Dbkey, dbinfo.Hostname, dbinfo.Dbname, dbinfo.Username, dbinfo.Password, int(dbinfo.Port)
}

func main() {
}
