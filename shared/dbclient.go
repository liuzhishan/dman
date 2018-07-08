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
func CGetDbinfo(dbkey, appkey string) (status int, dbkey1, hostname, dbname, username, password *C.char, port int) {
	err, dbinfo := cmd.GetDbinfo(dbkey, appkey)
	if err != nil {
		return
	}

	return 1, C.CString(dbinfo.Dbkey), C.CString(dbinfo.Hostname), C.CString(dbinfo.Dbname), C.CString(dbinfo.Username), C.CString(dbinfo.Password), int(dbinfo.Port)
}

func main() {
}
