//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cmd

import (
	"os"

	"github.com/IBM/k8s-volume-utils/pkg/kubeutils"
	"github.com/spf13/cobra"
)

// pvcCmd represents the pvc command
var pvcCmd = &cobra.Command{
	Use:   "pvc",
	Short: "rename PVC object in kubernetes clusterd",
	Long: `
Examples:
  k8s-volume-utils pvc "name" "targetName" "targetNamespace", "namespace"
  k8s-volume-utils pvc "name" "targetName" "targetNamespace"
  k8s-volume-utils pvc "name" "targetName"
Paramethers:
  name: pvc name to be renamed
  targetName: pvc name to be renamed to
  targetNamespace: namespace of new pvc
  namespace: namespace of the source pvc

`,
	Args: cobra.RangeArgs(2, 4),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(args) == 4 {
			err = kubeutils.RenamePVC(args[0], args[1], args[2], args[3])

		}
		if len(args) == 3 {
			err = kubeutils.RenamePVC(args[0], args[1], args[2], "")

		}
		if len(args) == 2 {
			err = kubeutils.RenamePVC(args[0], args[1], "", "")

		}
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pvcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pvcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pvcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
