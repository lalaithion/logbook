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
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/achew22/logbook/config"
	"github.com/achew22/logbook/parser"
	"github.com/achew22/logbook/templater"
)

var (
	nameOverride = flag.String("name_override", "", "Overrides the name of the user in the heading. Example --name_override=\"Joe Armstrong\"")
	dateOverride = flag.String("date_override", "", "Overrides the current date taking the form \"yyyy-mm-dd\". Example --date_override=1941-12-07")
)

func main() {
	flag.Parse()

	c := &config.Config{
		Name:    *nameOverride,
		LogPath: os.ExpandEnv("${HOME}/logbook"),
	}
	p := parser.New(c)

	if *nameOverride == "" {
		c.Name = "Andrew Allen"
	}

	var today parser.Date
	if *dateOverride == "" {
		today = parser.TimeToDate(time.Now())
	} else {
		var err error
		today, err = parser.YmdToDate(*dateOverride)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --date_override provided. %s", err)
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stderr, "Writing log entry for %s\n", today.ToYmd())

	if _, err := os.Stat(c.LogPath); err != nil {
		err := os.MkdirAll(c.LogPath, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Log path %s does not exist and could not be created\n", c.LogPath)
			os.Exit(1)
		}
	}

	todayPath := filepath.Join(c.LogPath, today.ToYmd()+".md")

	if _, err := os.Stat(todayPath); err == nil {
		fmt.Fprintf(os.Stderr, "A file already exists by the name %s\n", todayPath)
		os.Exit(1)
	}

	out, err := os.Create(todayPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create a log entry named %s.\nErr: %v", todayPath, err)
		os.Exit(1)
	}

	text := templater.Print(c, p.Parse(), today)

	out.Write([]byte(text))
	if err := out.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
