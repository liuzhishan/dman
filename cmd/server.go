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
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type (
	Server struct {
		addr       string
		httpServer *http.Server
		router     *httprouter.Router
	}
)

func mainHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func applyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	dbkey := r.FormValue("dbkey")
	appkey := r.FormValue("appkey")
	workername := r.FormValue("workername")
	info := r.FormValue("info")

	if isApplyExists(dbkey, appkey) {
		writeJSON(w, map[string]interface{}{"isSuccess": true, "message": "already applied"})
		return
	}

	err := insertApply(dbkey, appkey, workername, info)
	if err != nil {
		logInfo(err.Error())
		writeJSON(w, map[string]interface{}{"isSuccess": false, "message": "insert error!!!"})
		return
	}

	writeJSON(w, map[string]interface{}{"isSuccess": true, "message": "success, wait for admin"})
}

func checkHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	dbkey := r.FormValue("dbkey")
	appkey := r.FormValue("appkey")
	pub := []byte(r.FormValue("pub"))

	isSuccess, key := checkApply(dbkey, appkey)
	enkey, err := RsaEncrypt([]byte(key), pub)
	if err != nil {
		logInfo(err.Error())
		writeJSON(w, map[string]interface{}{"isSuccess": false, "key": ""})
	} else {
		writeJSON(w, map[string]interface{}{"isSuccess": isSuccess, "key": enkey})
	}
}

func writeJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		log.Println(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func NewServer() (*Server, error) {
	router := httprouter.New()
	router.GET("/", mainHandler)
	router.POST("/apply", applyHandler)
	router.POST("/check", checkHandler)

	srv := &Server{
		addr: fmt.Sprintf("%s:%d", viper.GetString("server.addr"), viper.GetInt("server.port")),
		httpServer: &http.Server{
			ReadTimeout:       time.Minute * 5,
			ReadHeaderTimeout: time.Minute * 2,
			IdleTimeout:       time.Minute * 5,
		},
		router: router,
	}

	return srv, nil
}

func (srv *Server) Serve() error {
	log.Printf("server on %v\n", srv.addr)

	err := http.ListenAndServe(srv.addr, logRequest(srv.router))
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) Close() error {
	if err := srv.httpServer.Close(); err != nil {
		return nil
	}

	return nil
}

func startServer() {
	srv, err := NewServer()
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("start server")
	err = srv.Serve()
	if err != nil {
		log.Println(err)
	}
}

func getBaseDir() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}

	base := dir[:strings.Index(dir, "dman")+len("dman")]
	return base
}

func main() {
	startServer()
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
