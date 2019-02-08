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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// getConfigCmd represents the config command
var getConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configuration", "settings", "cfg", "cfig", "cf"},
	Args:    cobra.NoArgs,
	Short:   "Prints out current configuration file contents",
	Long: `Prints used configuration (flags and/or environment variables)
within a configuration file.

	If using the default configuration file:
		./natscl get config
 	Or if using a custom configuration file:
		./natscl get config --config $HOME/envs/dev/.natscl.yaml`,

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

		bs, err := afero.ReadFile(afero.NewOsFs(), filename)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(bs))
		return nil
	},
}

func init() {
	getCmd.AddCommand(getConfigCmd)
}
