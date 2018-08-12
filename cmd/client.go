// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// Apply a dbkey for an appkey.
func Apply(dbkey, appkey, workername, info string) bool {
	reqUrl := fmt.Sprintf("%s/apply?dbkey=%s&appkey=%s", getBaseUrl(), dbkey, appkey)
	data := url.Values{"workername": {workername}, "info": {info}}
	resp, err := http.PostForm(reqUrl, data)
	if err != nil {
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return false
	}

	v, ok := body["isSuccess"]
	if ok {
		return v.(bool)
	}
	return false
}

func Check(dbkey, appkey string) bool {
	privateKey := getPrivateKey()
	publicKey := getPublicKey()

	reqUrl := fmt.Sprintf("%s/check?dbkey=%s&appkey=%s", getBaseUrl(), dbkey, appkey)
	data := url.Values{"pub": {string(publicKey)}}
	resp, err := http.PostForm(reqUrl, data)
	if err != nil {
		logInfo(err.Error())
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logInfo(err.Error())
		return false
	}

	jsonObj, err := gabs.ParseJSON(bodyBytes)
	if err != nil {
		logInfo(err.Error())
		return false
	}

	secretkey, ok := jsonObj.Path("key").Data().(string)
	if ok {
		decoded, err := hex.DecodeString(secretkey)
		if err != nil {
			logInfo(err.Error())
			return false
		}

		dekey, err := RsaDecrypt(decoded, privateKey)
		if err != nil {
			logInfo(err.Error())
			logInfo(string(secretkey))
			return false
		}
		storeKey(concatDbkeyAppkey(dbkey, appkey), string(dekey))
		return true
	}

	logInfo("%t", jsonObj.Path("isSuccess").Data().(bool))
	return false
}

func storeKey(k, v string) {
	fileKey := ".key.json"
	f, _ := ioutil.ReadFile(fileKey)
	data, err := gabs.ParseJSON(f)
	if err != nil {
		logInfo(err.Error())
		logInfo("new file")
		data = gabs.New()
	}

	data.Set(v, k)
	err = ioutil.WriteFile(fmt.Sprintf("%s.bak", fileKey), []byte(data.StringIndent("", " ")), 0644)
	if err != nil {
		logInfo(err.Error())
	} else {
		copyFile(fmt.Sprintf("%s.bak", fileKey), fileKey)
	}
}

func getKeyFromFile(k string) string {
	fileKey := ".key.json"
	f, _ := ioutil.ReadFile(fileKey)
	data, err := gabs.ParseJSON(f)
	if err != nil {
		logInfo(err.Error())
		return ""
	}

	m, ok := data.Data().(map[string]interface{})
	if ok {
		if v, ok := m[k].(string); ok {
			return v
		}
	}

	return ""
}

func concatDbkeyAppkey(dbkey, appkey string) string {
	return fmt.Sprintf("%s|%s", dbkey, appkey)
}

func GetDbinfo(dbkey, appkey string) (err error, dbinfo Dbinfo) {
	secretKey := getKeyFromFile(concatDbkeyAppkey(dbkey, appkey))
	privateKey := getPrivateKey()
	publicKey := getPublicKey()

	vcode := genVerifyCode(secretKey)

	reqUrl := fmt.Sprintf("%s/getDbinfo?dbkey=%s&appkey=%s", getBaseUrl(), dbkey, appkey)
	data := url.Values{"pub": {string(publicKey)}, "vcode": {vcode}}
	err, tmpInfo := sendPost(reqUrl, data)
	if err != nil {
		logInfo(err.Error())
		return err, dbinfo
	}

	res, _ := gabs.Consume(tmpInfo)
	content, ok := res.Path("content").Data().(string)
	if ok {
		decoded, err := hex.DecodeString(content)
		if err != nil {
			logInfo(err.Error())
			return err, dbinfo
		}

		decryptedInfo, err := RsaDecrypt(decoded, privateKey)
		if err != nil {
			logInfo(err.Error())
			return err, dbinfo
		}
		obj, _ := gabs.ParseJSON(decryptedInfo)
		dbinfo.Dbkey = obj.Path("dbkey").Data().(string)
		dbinfo.Hostname = obj.Path("hostname").Data().(string)
		dbinfo.Dbname = obj.Path("dbname").Data().(string)
		dbinfo.Username = obj.Path("username").Data().(string)
		dbinfo.Password = obj.Path("password").Data().(string)
		dbinfo.Port = int32(obj.Path("port").Data().(float64))

		logInfo("success, hostname: %s, dbname: %s, port: %d", dbinfo.Hostname, dbinfo.Dbname, dbinfo.Port)
		return nil, dbinfo
	}

	logInfo("error, not conent")
	return errors.New("no content"), dbinfo
}

func GetDbConnection(dbkey string, appkey string) (error, *sql.DB) {
	var db *sql.DB

	err, dbinfo := GetDbinfo(dbkey, appkey)
	if err != nil {
		logInfo(err.Error())
		return err, db
	}

	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", dbinfo.Username, dbinfo.Password, dbinfo.Hostname,
		dbinfo.Port, dbinfo.Dbname)
	err, db = openDb(dbUrl)

	return err, db
}

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client",
	Long:  "client",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply [dbkey] [appkey] [workername] [info]",
	Short: "apply",
	Long:  "apply dbkey for appkey",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		workername, info := "", ""
		if len(args) >= 3 {
			workername = args[2]
		}
		if len(args) >= 4 {
			info = args[3]
		}

		Apply(args[0], args[1], workername, info)
	},
}

var checkCmd = &cobra.Command{
	Use:   "check [dbkey] [appkey]",
	Short: "check",
	Long:  "check dbkey for appkey, if success, get key",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		Check(args[0], args[1])
	},
}

var getDbinfoCmd = &cobra.Command{
	Use:   "getDbinfo [dbkey] [appkey]",
	Short: "getDbinfo for dbkey and appkey",
	Long:  "getDbinfo for dbkey and appkey",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		GetDbinfo(args[0], args[1])
	},
}

var getDbConnectionCmd = &cobra.Command{
	Use:   "getDbConnection [dbkey] [appkey]",
	Short: "open db for dbkey and appkey",
	Long:  "open db for dbkey and appkey",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		GetDbConnection(args[0], args[1])
	},
}

func init() {
	clientCmd.AddCommand(applyCmd)
	clientCmd.AddCommand(checkCmd)
	clientCmd.AddCommand(getDbinfoCmd)
	clientCmd.AddCommand(getDbConnectionCmd)

	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	if _, err := os.Stat(".private.pem"); os.IsNotExist(err) {
		GenRsaKey()
	}
}
