/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/IBM/k8s-volume-utils/pkg/kubeutils"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// pvCmd represents the pv command
var pvCmd = &cobra.Command{
	Use:   "pv",
	Short: "help to reuse a pv",
	Long: `
k8s-volume-utils pv retain "pvName": change pv reclaim policy to retain
k8s-volume-utils pv resue "pvName": change pv status to available so that it is ready to be bound to pvc

`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "retain" {
			err := kubeutils.RetainPV(args[1])
			if err != nil {
				os.Exit(1)
			}
		} else if args[0] == "reuse" {
			err := kubeutils.ReusePV(args[1])
			if err != nil {
				os.Exit(1)
			}

		} else {
			klog.Error("invalid argument. should be retain or reuse")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pvCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
