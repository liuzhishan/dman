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
	"fmt"

	"github.com/spf13/cobra"
)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "admin",
	Long:  "admin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("admin called")
	},
}

var insertDbinfoCmd = &cobra.Command{
	Use:   "insertDbinfo [filename]",
	Short: "insert db info from json file",
	Long:  "insert db info from json file",
	Run: func(cmd *cobra.Command, args []string) {
		dbinfos := readDbinfoFromJson(args[0])
		for _, v := range dbinfos {
			insertDbinfo(v)
		}
	},
}

var showApplyCmd = &cobra.Command{
	Use:   "showApply [pattern]",
	Short: "show all apply",
	Long:  "show all apply",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		showApply(pattern)
	},
}

var approveApplyCmd = &cobra.Command{
	Use:   "approveApply [applyid,username] [applyid1,username1] ...",
	Short: "approve apply with username",
	Long:  "approve apply with username",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		approveApply(args)
	},
}

func init() {
	adminCmd.AddCommand(insertDbinfoCmd)
	adminCmd.AddCommand(showApplyCmd)
	adminCmd.AddCommand(approveApplyCmd)

	rootCmd.AddCommand(adminCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// adminCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// adminCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
