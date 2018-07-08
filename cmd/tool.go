package cmd

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	_ "reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	serverAddr = "localhost"
	serverPort = 8072
	db         *sql.DB
)

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
	return fmt.Sprintf("http://%s:%d", serverAddr, serverPort)
}

func openDb(dbUrl string) (err error, db *sql.DB) {
	db, err = sql.Open("mysql", dbUrl)
	err = db.Ping()
	if err != nil {
		return err, db
	}

	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	logInfo("get db connection success!")
	return nil, db
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

func isUsernameExists(username string) bool {
	s := "select count(1) from dbaccount where username = ?"
	params := []interface{}{username}
	err, rows := getSqlResult(s, params, []interface{}{0})
	if err != nil {
		logInfo(err.Error())
		return false
	}

	return rows[0][0].(int64) > 0
}

func approveApply(applys []string) {
	for _, s := range applys {
		splitted := strings.Split(s, ",")
		if len(splitted) == 2 {
			applyidStr, username := splitted[0], splitted[1]
			applyid, err := strconv.Atoi(applyidStr)
			if err != nil {
				logInfo("error! applyid must be int, %s", err.Error())
				continue
			}
			if !isUsernameExists(username) {
				logInfo("error! username not exists!!!")
				continue
			}
			s := "update user_apply set status = 1, secretkey = ?, username = ? where applyid = ?"
			key := randStr(32)
			params := []interface{}{key, username, applyid}
			err = executeSql(s, params)
			if err != nil {
				logInfo(err.Error())
			} else {
				logInfo("approve success, applyid = %d, username = %s", applyid, username)
			}
		} else {
			logInfo("wrong format, must be applyid,username")
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

func sendGet(reqUrl string) (err error, res interface{}) {
	return
}

func sendPost(reqUrl string, data url.Values) (err error, res interface{}) {
	resp, err := http.PostForm(reqUrl, data)
	if err != nil {
		logInfo(err.Error())
		return err, res
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logInfo(err.Error())
		return err, res
	}

	jsonObj, err := gabs.ParseJSON(bodyBytes)
	if err != nil {
		logInfo(err.Error())
		return err, res
	}

	return nil, jsonObj.Data()
}

func genVerifyCode(secretKey string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s%d", secretKey, time.Now().Unix()/30)))
	x := binary.BigEndian.Uint64(h.Sum(nil)[:8])

	res := fmt.Sprintf("%06d", x%1000000)
	return res
}

func checkAccess(dbkey, appkey, vcode string) (isSuccess bool, username string, info string) {
	s := "select status, coalesce(secretkey, ''), coalesce(username, '') from user_apply where dbkey = ? and appkey = ?"
	params := []interface{}{dbkey, appkey}

	err, rows := getSqlResult(s, params, []interface{}{0, "", ""})
	if err != nil {
		return false, "", "error when get key"
	} else if len(rows) == 0 {
		return false, "", "no such dbkey for appkey"
	} else if rows[0][0].(int64) == 0 {
		return false, "", "not ready, waiting for approve, please contact dba"
	} else if rows[0][0].(int64) == 1 {
		secretKey := string(rows[0][1].([]byte))
		username := string(rows[0][2].([]byte))
		if genVerifyCode(secretKey) != vcode {
			return false, "", "wrong verify code"
		} else {
			return true, username, ""
		}
	} else {
		return false, "", ""
	}
}

func getDbinfoByDbkey(dbkey string, username string) (err error, dbinfo map[string]interface{}) {
	s := "select hostname, dbname, password, port from dbaccount where dbkey = ? and username = ?"
	params := []interface{}{dbkey, username}
	err, rows := getSqlResult(s, params, []interface{}{"", "", "", 0})
	if err != nil {
		return err, dbinfo
	}

	dbinfo = map[string]interface{}{"dbkey": dbkey, "username": username}
	dbinfo["hostname"] = string(rows[0][0].([]byte))
	dbinfo["dbname"] = string(rows[0][1].([]byte))
	dbinfo["password"] = string(rows[0][2].([]byte))
	dbinfo["port"] = rows[0][3].(int64)

	logInfo("success, hostname: %s, dbname: %s, port: %d", dbinfo["hostname"], dbinfo["dbname"], dbinfo["port"])
	return nil, dbinfo
}
