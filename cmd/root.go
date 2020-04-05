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
	"flag"
	"fmt"
	"os"

	"github.com/IBM/k8s-volume-utils/pkg/kubeutils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-volume-utils",
	Short: "it is used to rename PVC object or help to reuse pv",
	Long: `
k8s-volume-utils pvc
   It is used to rename a PVC object in kubernetes cluster.
   It will delete original PVC and create new PVC referring to same PV.
   New PVC can be in another namespace.
k8s-volume-utils pv
   It can help to reuse a pv
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	klog.InitFlags(nil)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	flags := rootCmd.PersistentFlags()

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(rootCmd.PersistentFlags())

	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	var namespace string
	var err error
	namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		klog.Errorf("fail to get default namespace: %s", err.Error())
	}

	client, e := f.KubernetesClientSet()
	if e != nil {
		klog.Errorf("fail to get kubeclient: %s", e.Error())
	}
	kubeutils.InitKube(client, namespace)
	klog.Infof("kubeclient initialized. default namespace: %s", namespace)
}
