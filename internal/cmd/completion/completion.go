package completion

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/riywo/loginshell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cmd/options"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/utils/flags"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions

		Shell string
	}
)

func NewCmd(opts *options.CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *options.CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "completion",
		Short:         "Output shell completion code",
		SilenceErrors: true,
		SilenceUsage:  true,
		Example:       usageExample,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *Cmd) Run(ctx context.Context) error {
	steps := []func(context.Context) error{
		x.survey,
		x.log,
		x.validate,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *Cmd) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *Cmd) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *Cmd) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *Cmd) run(ctx context.Context) error {
	switch x.opts.Shell {
	case "bash":
		return x.runCmdBash(x.opts.Out, "")
	case "zsh":
		return x.runCmdZsh(x.opts.Out, "")
	default:
		return errors.Internal(fmt.Errorf("unsupported shell: %s", x.opts.Shell))
	}
}

func (x *Cmd) runCmdBash(out io.Writer, boilerplate string) error {
	if len(boilerplate) == 0 {
		boilerplate = defaultBoilerplate
	}

	if _, err := out.Write([]byte(boilerplate)); err != nil {
		return err
	}

	return x.cmd.Root().GenBashCompletion(out)
}

func (x *Cmd) runCmdZsh(out io.Writer, boilerplate string) error {
	if len(boilerplate) == 0 {
		boilerplate = defaultBoilerplate
	}

	if _, err := out.Write([]byte(boilerplate)); err != nil {
		return err
	}

	// TODO(liran): Fixes https://github.com/spf13/cobra/issues/881.
	io.WriteString(out, fmt.Sprintf("\ncompdef _%s %s\n",
		x.cmd.Root().Name(),
		x.cmd.Root().Name()))

	return x.cmd.Root().GenZshCompletion(out)
}

const defaultBoilerplate = `
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

const usageExample = `
Outputs shell completion for the given shell (bash or zsh).
This depends on the bash-completion binary. 
Example installation instructions:

macOS:
	$ brew install bash-completion
	$ source $(brew --prefix)/etc/bash_completion
	$ mkdir -p ~/.spotinst
	$ spotctl completion > ~/.spotinst/spotctl-completion
	$ source ~/.spotinst/spotctl-completion

Ubuntu:
	$ apt-get install bash-completion
	$ source /etc/bash-completion
	$ mkdir -p ~/.spotinst
	$ spotctl completion > ~/.spotinst/spotctl-completion
	$ source ~/.spotinst/spotctl-completion

Additionally, you may want to output the completion to a file and source in your shell rcfile.`

func (x *CmdOptions) Init(fs *pflag.FlagSet, opts *options.CommonOptions) {
	x.initFlags(fs)
	x.initDefaults(opts)
}

func (x *CmdOptions) initDefaults(opts *options.CommonOptions) {
	x.CommonOptions = opts

	if shell, err := loginshell.Shell(); err == nil {
		parts := strings.Split(shell, "/")
		if len(parts) > 0 {
			x.Shell = strings.ToLower(parts[len(parts)-1:][0])
		}
	}
}

func (x *CmdOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.Shell, "shell", x.Shell, "name of the shell")
}

func (x *CmdOptions) Validate() error {
	return x.CommonOptions.Validate()
}
