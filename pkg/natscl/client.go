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

package natscl

import (
	"encoding/json"
	"errors"
	"github.com/nats-io/go-nats-streaming"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
)

// MusBin panics if err is non-nil otherwise returns the remaining values.
func must(a interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return a
}

type Message struct {
	Sequence    uint64 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Subject     string `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	Data        string `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	Timestamp   int64  `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Redelivered bool   `protobuf:"varint,6,opt,name=redelivered,proto3" json:"redelivered,omitempty"`
}

type Client struct {
	cc        stan.Conn
	mu        sync.Mutex
	subs      []stan.Subscription
	validSubs uint64
	Logger    *log.Logger
}

func New(clusterID string, option ...stan.Option) (*Client, error) {
	cli := new(Client)
	cli.Logger = log.New(os.Stderr, "", log.LstdFlags)

	cli.Logger.Println("connecting...")
	cc, err := stan.Connect(clusterID, filepath.Base(must(os.Executable()).(string))+"-"+strconv.Itoa(os.Getpid()), option...)
	if err != nil {
		return nil, errors.New("unable to connect to cluster: " + err.Error())
	}
	cli.Logger.Println("connected to", cc.NatsConn().ConnectedUrl(), "["+clusterID+"]")
	cli.cc = cc
	return cli, nil
}
func (m *Client) Write(r io.Reader, subject string) error {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return m.cc.Publish(subject, bs)
}
func (m *Client) Read(w io.Writer, readCount uint64, subject string, group string, filters []*regexp.Regexp, option ...stan.SubscriptionOption) error {
	m.mu.Lock()
	m.Logger.Println("subscribing", subject)

	hasReadLimit := readCount > 0
	subs, err := m.cc.QueueSubscribe(subject, group, func(msg *stan.Msg) {
		msg2 := &Message{
			Sequence:    msg.Sequence,
			Subject:     msg.Subject,
			Data:        string(msg.Data),
			Timestamp:   msg.Timestamp,
			Redelivered: msg.Redelivered,
		}

		for _, filter := range filters {
			if !filter.Match(msg.Data) {
				return
			}
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		m.mu.Lock()

		if err := enc.Encode(msg2); err != nil {
			m.Logger.Fatalln("error encoding message:", err)
		}
		if err := msg.Ack(); err != nil {
			m.Logger.Fatalln("error acknowledging message:", err)
		}
		readCount--
		if hasReadLimit && readCount == 0 {
			atomic.AddUint64(&m.validSubs, ^uint64(0))
			if err := msg.Sub.Close(); err != nil {
				m.Logger.Println("error closing subscription: read limit reached:", msg.Subject)
			} else {
				m.Logger.Println("closed subscription: read limit reached:", msg.Subject)
			}
		}
		m.mu.Unlock()
	}, option...)
	m.subs = append(m.subs, subs)
	atomic.AddUint64(&m.validSubs, 1)
	m.mu.Unlock()
	return err
}
func (m *Client) ValidSubscriptions() uint64 {
	return atomic.LoadUint64(&m.validSubs)
}
func (m *Client) Close() (err error) {
	m.mu.Lock()
	for _, sub := range m.subs {
		if sub.IsValid() {
			m.Logger.Println("closing subscription:", sub)
			if err = sub.Close(); err != nil {
				m.mu.Unlock()
				return err
			}
		}
	}

	m.Logger.Println("closing connection to", m.cc.NatsConn().ConnectedUrl())
	err = m.cc.Close()
	m.mu.Unlock()
	return
}
