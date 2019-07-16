package completion

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/riywo/loginshell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cmd/options"
	"github.com/spotinst/spotinst-cli/internal/errors"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions
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
	steps := []func(context.Context) error{x.survey, x.validate, x.run}

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

func (x *Cmd) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *Cmd) run(ctx context.Context) error {
	shell, err := loginshell.Shell()
	if err != nil {
		return errors.Internal(err)
	}

	parts := strings.Split(shell, "/")
	if len(parts) > 0 {
		shell = strings.ToLower(parts[len(parts)-1:][0])
	}

	switch shell {
	case "bash":
		return x.runCmdBash(x.opts.Out, "")
	case "zsh":
		return x.runCmdZsh(x.opts.Out, "")
	default:
		return errors.Internal(fmt.Errorf("unsupported shell: %s", shell))
	}
}

func (x *Cmd) runCmdBash(out io.Writer, boilerplate string) error {
	if len(boilerplate) == 0 {
		boilerplate = defaultBoilerplate
	}

	if _, err := out.Write([]byte(boilerplate)); err != nil {
		return err
	}

	return x.cmd.GenBashCompletion(out)
}

func (x *Cmd) runCmdZsh(out io.Writer, boilerplate string) error {
	zshHead := "#compdef spotinst\n"
	out.Write([]byte(zshHead))

	if len(boilerplate) == 0 {
		boilerplate = defaultBoilerplate
	}

	if _, err := out.Write([]byte(boilerplate)); err != nil {
		return err
	}

	zshInitialization := `
__spotinst_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}

__spotinst_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift

		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__spotinst_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}

__spotinst_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?

	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}

__spotinst_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__spotinst_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}

__spotinst_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}

__spotinst_filedir() {
	local RET OLD_IFS w qw

	__spotinst_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi

	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"

	IFS="," __spotinst_debug "RET=${RET[@]} len=${#RET[@]}"

	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__spotinst_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}

__spotinst_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
	printf %q "$1"
    fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi

__spotinst_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__spotinst_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__spotinst_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__spotinst_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__spotinst_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__spotinst_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__spotinst_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	x.cmd.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}

__spotinst_bash_source <(__spotinst_convert_bash_to_zsh)
_complete spotinst 2>/dev/null
`
	out.Write([]byte(zshTail))
	return nil
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
# Installing bash completion on macOS

$ 
$ spotinst completion > $(brew --prefix)/etc/bash_completion.d/spotinst


# Installing bash completion on Linux
## If bash-completion is not installed on Linux, please install the 'bash-completion' package
## via your distribution's package manager.
## Load the kubectl completion code for bash into the current shell
	source <(kubectl completion bash)
## Write bash completion code to a file and source if from .bash_profile
	kubectl completion bash > ~/.kube/completion.bash.inc
	printf "
	  # Kubectl shell completion
	  source '$HOME/.kube/completion.bash.inc'
	  " >> $HOME/.bash_profile
	source $HOME/.bash_profile

# Load the kubectl completion code for zsh[1] into the current shell
	source <(kubectl completion zsh)
# Set the kubectl completion code for zsh[1] to autoload on startup
	kubectl completion zsh > "${fpath[1]}/_kubectl"
`

func (x *CmdOptions) Init(flags *pflag.FlagSet, opts *options.CommonOptions) {
	x.CommonOptions = opts
}

func (x *CmdOptions) Validate() error {
	return x.CommonOptions.Validate()
}
