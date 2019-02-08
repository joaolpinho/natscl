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
	"bufio"
	"bytes"
	"fmt"
	"github.com/joaolpinho/natscl/pkg/natscl"
	"github.com/nats-io/go-nats-streaming"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
)

// saveMessagesCmd represents the messages command
var saveMessagesCmd = &cobra.Command{
	Use:   "messages",
	Aliases: []string{"message", "msg"},
	Short: "Publish to NATS topics.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log.SetOutput(os.Stderr)

		cluster, _ := cmd.Flags().GetString("cluster")
		url, _ := cmd.Flags().GetString("url")
		cli, err := natscl.New(cluster, stan.NatsURL(url), stan.SetConnectionLostHandler(func(conn stan.Conn, e error) {
			fmt.Println("connection lost...")
		}))
		if err != nil {
			return err
		}
		defer func () {
			log.Println("closing connections...")
			cli.Close()
			log.Println("disconnected.")
		}()


		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		reader := bufio.NewReader(os.Stdin)
		for {
			select {

			case <-c:
				return nil
			default:
				if line,err := reader.ReadString('\n'); err != nil {
					if err != io.EOF {
						log.Println("error reading input:", err)
						return err
					}
					return nil
				} else if line == "\n" {
					continue
				} else if err := cli.Write(bytes.NewBufferString(strings.Trim(line, " \n")), args[0]); err != nil {
					log.Println("error writing message:", err)
					return err
				}
			}

		}
	},
}

func init() {
	saveCmd.AddCommand(saveMessagesCmd)
}
