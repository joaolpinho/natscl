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

package datefmt

import (
	"strings"
	"time"
)

var words = map[string]string{
	"weekday": "Monday",
	"DD":      "02",
	"D":       "_2",
	"MMM":     "Jan",
	"MM":      "01",
	"YYYY":    "2006",
	"YY":      "06",

	"hh":  "15",
	"mm":  "04",
	"ss":  "05",
	"s":   "000000000",
	"TZD": "Z07:00",

	"time12": "3:04:05 pm",
	"time":   "15:04:05.00",
	"time24": "15:04:05",
}

func Normalize(fmt string) string {
	for k, v := range words {
		fmt = strings.Replace(fmt, k, v, 1)
	}
	return fmt
}

func Format(date time.Time, layout string) string {
	return date.Format(Normalize(layout))
}

func Parse(date string, layout string) (time.Time, error) {
	return time.Parse(Normalize(layout), date)
}

func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(Normalize(layout), value, loc)
}

var layouts = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	"2006-01-02T15:04:05.000000000Z07:00",
	"2006-01-02",
	"2006-01-02 15:04:05",
}

func ParseWithoutLayout(date string) (t time.Time, err error) {
	for _, layout := range layouts {
		t, err = time.Parse(layout, date)
		if err == nil {
			break
		}
	}
	return
}
