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
	"github.com/joaolpinho/natscl/pkg/datefmt"
	"github.com/joaolpinho/natscl/pkg/natscl"
	"github.com/nats-io/go-nats-streaming"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"regexp"
	"time"
)

// getMessagesCmd represents the messages command
var getMessagesCmd = &cobra.Command{
	Use:     "messages",
	Aliases: []string{"message", "msg"},
	Short:   "Subscribes to NATS topics.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		log := log.New(cmd.OutOrStderr(), "", log.LstdFlags)

		cli, err := natscl.New(must(flags.GetString("cluster")).(string),
			stan.NatsURL(must(flags.GetString("url")).(string)),
			stan.SetConnectionLostHandler(func(conn stan.Conn, e error) {
				fmt.Println("connection lost:", e)
			}))
		if err != nil {
			return err
		}

		cli.Logger = log

		ops := []stan.SubscriptionOption{stan.SetManualAckMode(), stan.MaxInflight(1)}
		if since, _ := flags.GetString("since"); since != "" {
			dur, err := time.ParseDuration(since)
			if err != nil {
				t, err := datefmt.ParseWithoutLayout(since)
				if err != nil {
					return err
				}
				ops = append(ops, stan.StartAtTime(t))
			} else {
				ops = append(ops, stan.StartAtTimeDelta(dur))
			}
		} else if sq, _ := flags.GetUint64("seq"); sq > 0 {
			ops = append(ops, stan.StartAtSequence(sq))
		} else if last, _ := flags.GetBool("last"); last {
			ops = append(ops, stan.StartWithLastReceived())
		} else {
			ops = append(ops, stan.DeliverAllAvailable())
		}

		durable, _ := cmd.Flags().GetString("durable")
		if durable != "" {
			ops = append(ops, stan.DurableName(durable))
		}

		filters := make([]*regexp.Regexp, 0, 5)
		for _, expr := range must(flags.GetStringSlice("filter")).([]string) {
			rxp, err := regexp.Compile(expr)
			if err != nil {
				return err
			}
			filters = append(filters, rxp)
		}

		readCount, _ := cmd.Flags().GetUint64("read-count")
		group, _ := cmd.Flags().GetString("group")
		for _, subject := range args {
			if err := cli.Read(cmd.OutOrStdout(), readCount, subject, group, filters, ops...); err != nil {
				return err
			}
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			select {
			case <-c:
				log.Println("closing connections...")
				if err := cli.Close(); err != nil {
					log.Println("error closing connection:", err)
				}
				log.Println("disconnected.")
				return nil
			default:
				if cli.ValidSubscriptions() == 0 {
					log.Println("all subscriptions are closed")
					log.Println("closing connections...")
					if err := cli.Close(); err != nil {
						log.Println("error closing connection:", err)
					}
					log.Println("disconnected.")
					return nil
				}
			}
		}
	},
}

func init() {
	getCmd.AddCommand(getMessagesCmd)
	getMessagesCmd.Flags().StringP("durable", "d", "", "Durable queue name")
	getMessagesCmd.Flags().StringP("group", "g", "", "Subscription group name")
	getMessagesCmd.Flags().StringP("since", "t", "", "Start reading from messages published since passed time")
	getMessagesCmd.Flags().BoolP("last", "l", false, "Start reading from last received")
	getMessagesCmd.Flags().Uint64P("seq", "s", 0, "Start reading from sequence no.")
	getMessagesCmd.Flags().StringSliceP("filter", "f", []string{}, "Filter read messages output.")
	getMessagesCmd.Flags().Uint64P("read-count", "r", 0, "Start reading from sequence no.")
}
