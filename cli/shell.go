// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cli

import (
	"fmt"
	"io"
	"strings"

	"cmk/config"

	"github.com/mattn/go-shellwords"

	"github.com/chzyer/readline"
	"os"
)

var completer *autoCompleter
var shell *readline.Instance

// ExecShell starts a shell
func ExecShell(sysArgs []string) {
	cfg := config.NewConfig()
	completer = &autoCompleter{
		Config: cfg,
	}

	if len(sysArgs) > 0 {
		err := ExecCmd(cfg, sysArgs)
		if err != nil {
			fmt.Println("🙈 Error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	shell, err := readline.NewEx(&readline.Config{
		Prompt:            cfg.GetPrompt(),
		HistoryFile:       cfg.HistoryFile,
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		VimMode:           false,
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
	})

	if err != nil {
		panic(err)
	}
	defer shell.Close()

	cfg.HasShell = true
	cfg.PrintHeader()

	for {
		shell.SetPrompt(cfg.GetPrompt())
		line, err := shell.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if len(line) < 1 {
			continue
		}

		shellwords.ParseEnv = true
		parser := shellwords.NewParser()
		args, err := parser.Parse(line)
		if err != nil {
			fmt.Println("Failed to parse line:", err)
			continue
		}

		if parser.Position > 0 {
			line = fmt.Sprintf("shell %s %v", cfg.Name(), line)
			args = strings.Split(line, " ")
		}

		err = ExecCmd(cfg, args)
		if err != nil {
			fmt.Println("🙈 Error:", err)
		}
	}
}
