package cmd

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	_ "reflect"
	"runtime"
	"strconv"
	"strings"
)

var db *sql.DB

func logInfo(formating string, args ...interface{}) {
	filename, line, funcname := "???", 0, "???"
	pc, filename, line, ok := runtime.Caller(1)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()
		funcname = filepath.Ext(funcname)
		funcname = strings.TrimPrefix(funcname, ".")

		filename = filepath.Base(filename)
	}

	log.Printf("%s [%s] line %d: %s\n", filename, funcname, line, fmt.Sprintf(formating, args...))
}

func getBaseUrl() string {
	return fmt.Sprintf("http://%s:%d", viper.GetString("server.addr"), viper.GetInt("server.port"))
}

func openDb() *sql.DB {
	db, err := sql.Open("mysql", viper.GetString("db.url"))
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()
	logInfo("get db connection success! dbname: %s", viper.GetString("db.dbname"))
	return db
}

func getSqlResult(s string, params []interface{}, example []interface{}) (err error, rows [][]interface{}) {
	sqlRes, err := db.Query(s, params...)
	defer sqlRes.Close()
	if err != nil {
		logInfo(err.Error())
		return err, rows
	}

	dest := make([]interface{}, len(example))
	for i, _ := range example {
		dest[i] = &example[i]
	}
	rows = make([][]interface{}, 0)
	for sqlRes.Next() {
		err = sqlRes.Scan(dest...)
		if err != nil {
			logInfo(err.Error())
			return err, rows
		}

		tmp := make([]interface{}, len(example))
		copy(tmp, example)
		rows = append(rows, tmp)
	}

	return nil, rows
}

func executeSql(s string, params []interface{}) (err error) {
	statement, err := db.Prepare(s)
	if err != nil {
		logInfo(fmt.Sprintf("statement error: %s", err.Error()))
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(params...)
	if err != nil {
		logInfo(err.Error())
		return err
	}
	return nil
}

type Dbinfo struct {
	Dbkey    string `json:"dbkey"`
	Hostname string `json:"hostname"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int32  `json:"port"`
}

func (info Dbinfo) toString() string {
	bytes, err := json.Marshal(info)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	return string(bytes)
}

func readDbinfoFromJson(filename string) []Dbinfo {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	var c []Dbinfo
	json.Unmarshal(raw, &c)
	return c
}

func insertDbinfo(info Dbinfo) {
	s := "REPLACE INTO dbaccount (dbkey, hostname, dbname, port, username, password) VALUES(?, ?, ?, ?, ?, ?)"
	params := []interface{}{info.Dbkey, info.Hostname, info.Dbname, info.Port, info.Username, info.Password}
	executeSql(s, params)
}

func isApplyExists(dbkey string, appkey string) bool {
	s := "select count(1) from user_apply where dbkey = ? and appkey = ?"
	params := []interface{}{dbkey, appkey}

	err, rows := getSqlResult(s, params, []interface{}{1})
	if err != nil {
		logInfo(err.Error())
		return true
	}
	if len(rows) == 1 {
		return rows[0][0].(int64) > 0
	}

	return false
}

func insertApply(dbkey string, appkey string, workername string, info string) error {
	s := "REPLACE INTO user_apply (dbkey, appkey, workername, info, status) VALUES(?, ?, ?, ?, ?)"
	params := []interface{}{dbkey, appkey, workername, info, 0}
	return executeSql(s, params)
}

func checkApply(dbkey string, appkey string) (isSuccess bool, secretkey string) {
	s := "select coalesce(status, 0), coalesce(secretkey, '') from user_apply where dbkey = ? and appkey = ?"
	params := []interface{}{dbkey, appkey}

	err, rows := getSqlResult(s, params, []interface{}{1, ""})
	if err != nil || len(rows) == 0 {
		return false, ""
	}
	secretkey = string(rows[0][1].([]byte))
	return rows[0][0].(int64) == 1, secretkey
}

func showApply(pattern string) {
	s := "select applyid, status, dbkey, appkey, workername, info from user_apply where appkey like ?"
	params := []interface{}{fmt.Sprintf("%%%s%%", pattern)}
	err, rows := getSqlResult(s, params, []interface{}{0, 0, "", "", "", ""})
	if err != nil {
		logInfo(err.Error())
		return
	}

	logInfo(strings.Join([]string{"applyid", "status", "dbkey", "appkey", "workername", "info"}, "\t"))
	for _, row := range rows {
		logInfo("%d\t%d\t%s\t%s\t%s\t%s", row...)
	}
}

func approveApply(applyids string) {
	for _, applyid := range strings.Split(applyids, ",") {
		s := "update user_apply set status = 1, secretkey = ? where applyid = ?"
		key := randStr(32)
		x, err := strconv.Atoi(applyid)
		if err != nil {
			logInfo(err.Error())
			continue
		}
		params := []interface{}{key, x}
		err = executeSql(s, params)
		if err != nil {
			logInfo(err.Error())
		} else {
			logInfo("approve success, applyid = %d", x)
		}
	}
}

func copyFile(src string, dst string) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		logInfo(err.Error())
		return
	}

	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		logInfo(err.Error())
		return
	}
}

func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	return str[:len]
}
