// Copyright © 2019 João Lopes Pinho <joaolpinho@gmail.com>
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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// saveConfigCmd represents the config command
var saveConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configuration", "settings", "cfg", "cfig", "cf"},
	Args:    cobra.NoArgs,
	Short:   "Saves used flags to a configuration file",
	Long: `Persists used configuration (flags and/or environment variables)
within a configuration file:

	For examples for saving the cluster name and the NATS URL:
		./natscl save config --cluster custom-cluster --url nats://nats-io`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		filename, err := flags.GetString("config")
		if err != nil {
			return err
		}
		if filename == "" {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				return err
			}
			filename = home + "/.natscl.yaml"
		}
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Name != "config" && flag.Name != "help" {
				viper.Set(flag.Name, flag.Value.String())
			}
		})
		err = viper.WriteConfigAs(filename)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "written configuration to", filename)
		return nil
	},
}

func init() {
	saveCmd.AddCommand(saveConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// saveConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// saveConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
