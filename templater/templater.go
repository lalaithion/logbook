/**
 * Copyright 2018 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the license.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specified language governing permissions and
 * limitations under the License.
 */
package templater

import (
	"fmt"
	"os"
	"strings"

	"github.com/achew22/logbook/config"
	"github.com/achew22/logbook/parser"
)

type extractedEntry struct {
	originDate parser.Date
	text       string
}

func Print(c *config.Config, entries map[parser.Date]*parser.LogEntry, today parser.Date) string {
	buf := &strings.Builder{}
	forToday := []extractedEntry{}
	for _, entry := range entries {
		for date, lines := range entry.FutureReferences {
			if today.Equals(date) {
				for _, line := range lines {
					forToday = append(forToday, extractedEntry{
						date,
						line,
					})
				}
			}
		}
	}

	fmt.Fprintf(buf, "# %s - %s\n\n", c.Name, today.ToYmd())

	if len(forToday) > 0 {
		fmt.Fprintf(buf, "Reminders:\n")

		for _, e := range forToday {
			fmt.Fprintf(buf, "From %s: %s\n", e.originDate.ToYmd(), e.text)
		}
	}

	fmt.Fprintf(buf, "\n")

	fmt.Fprintf(os.Stderr, "Wrote file\n")
	return buf.String()
}
