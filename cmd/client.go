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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
)

func apply(dbkey, appkey, workername, info string) bool {
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

func check(dbkey, appkey string) (err error, res interface{}) {
	privateKey := getPrivateKey()
	publicKey := getPublicKey()

	reqUrl := fmt.Sprintf("%s/check?dbkey=%s&appkey=%s", getBaseUrl(), dbkey, appkey)
	data := url.Values{"pub": {string(publicKey)}}
	resp, err := http.PostForm(reqUrl, data)
	if err != nil {
		logInfo(err.Error())
		return err, res
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logInfo(err.Error())
		return
	}

	jsonObj, err := gabs.ParseJSON(bodyBytes)
	if err != nil {
		logInfo(err.Error())
		return err, jsonObj.Data()
	}

	secretkey, ok := jsonObj.Path("key").Data().(string)
	if ok {
		dekey, err := RsaDecrypt([]byte(secretkey), privateKey)
		if err != nil {
			logInfo(err.Error())
			logInfo(secretkey)
			return err, nil
		}
		logInfo(string(dekey))
		storeKey(fmt.Sprintf("%s|%s", dbkey, appkey), string(dekey))
	}
	return nil, jsonObj.Data()
}

func storeKey(k, v string) {
	fileKey := ".key.json"
	f, _ := ioutil.ReadFile(fileKey)
	data, err := gabs.ParseJSON(f)
	if err != nil {
		logInfo(err.Error())
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

	if v, ok := data.Path(k).Data().(string); ok {
		return v
	}
	return ""
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

		apply(args[0], args[1], workername, info)
	},
}

var checkCmd = &cobra.Command{
	Use:   "check [dbkey] [appkey]",
	Short: "check",
	Long:  "check dbkey for appkey, if success, get key",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		check(args[0], args[1])
	},
}

func init() {
	clientCmd.AddCommand(applyCmd)
	clientCmd.AddCommand(checkCmd)

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
