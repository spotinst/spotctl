package survey

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// Survey holds the configuration for the survey.
type Survey struct {
	// In, Out, and Err represent the respective data streams that the survey
	// may act upon. They are attached directly to each prompt.
	In  terminal.FileReader
	Out terminal.FileWriter
	Err io.Writer
}

// New returns a new survey interface.
func New(in io.Reader, out, err io.Writer) Interface {
	var surv Survey

	if v, ok := in.(terminal.FileReader); ok {
		surv.In = v
	}

	if v, ok := out.(terminal.FileWriter); ok {
		surv.Out = v
	}

	if v, ok := err.(terminal.FileWriter); ok {
		surv.Err = v
	}

	return &surv
}

func (x *Survey) InputString(input *Input) (string, error) {
	surveyOpts := []survey.AskOpt{
		survey.WithStdio(x.In, x.Out, x.Err),
	}

	answer := ""
	prompt := &survey.Input{
		Message: input.Message,
	}

	if input.Default != nil {
		prompt.Default = fmt.Sprintf("%v", input.Default)
	}

	if input.Help != "" {
		prompt.Help = input.Help
	}

	if input.Required {
		surveyOpts = append(surveyOpts, survey.WithValidator(survey.Required))
	}

	question := &survey.Question{
		Prompt: prompt,
	}

	if input.Validate != nil {
		question.Validate = survey.Validator(input.Validate)
	}

	if input.Transform != nil {
		question.Transform = survey.Transformer(input.Transform)
	}

	return answer, survey.Ask([]*survey.Question{question}, &answer, surveyOpts...)
}

func (x *Survey) InputInt64(input *Input) (int64, error) {
	answer, err := x.InputString(input)
	if err != nil {
		return 0, err
	}
	if answer == "" {
		return 0, nil
	}

	return strconv.ParseInt(answer, 10, 64)
}

func (x *Survey) InputFloat64(input *Input) (float64, error) {
	answer, err := x.InputString(input)
	if err != nil {
		return 0, err
	}
	if answer == "" {
		return 0, nil
	}

	return strconv.ParseFloat(answer, 64)
}

func (x *Survey) Password(input *Input) (string, error) {
	surveyOpts := []survey.AskOpt{
		survey.WithStdio(x.In, x.Out, x.Err),
	}

	answer := ""
	prompt := &survey.Password{
		Message: input.Message,
	}

	if input.Help != "" {
		prompt.Help = input.Help
	}

	if input.Required {
		surveyOpts = append(surveyOpts, survey.WithValidator(survey.Required))
	}

	question := &survey.Question{
		Prompt: prompt,
	}

	if input.Validate != nil {
		question.Validate = survey.Validator(input.Validate)
	}

	if input.Transform != nil {
		question.Transform = survey.Transformer(input.Transform)
	}

	return answer, survey.Ask([]*survey.Question{question}, &answer, surveyOpts...)
}

func (x *Survey) Confirm(input *Input) (bool, error) {
	surveyOpts := []survey.AskOpt{
		survey.WithStdio(x.In, x.Out, x.Err),
	}

	answer := false
	var err error

	if v, ok := input.Default.(string); ok {
		if answer, err = strconv.ParseBool(v); err != nil {
			return answer, err
		}
	}

	prompt := &survey.Confirm{
		Message: input.Message,
		Default: answer,
	}

	if input.Help != "" {
		prompt.Help = input.Help
	}

	return answer, survey.AskOne(prompt, &answer, surveyOpts...)
}

func (x *Survey) Select(input *Select) (string, error) {
	var selected string

	if len(input.Options) == 0 {
		return selected, nil
	}

	surveyOpts := []survey.AskOpt{
		survey.WithStdio(x.In, x.Out, x.Err),
	}

	options := make([]string, 0, len(input.Options))
	for _, opt := range input.Options {
		if v, ok := opt.(string); ok {
			options = append(options, v)
		}
	}

	prompt := &survey.Select{
		Message: input.Message,
		Options: options,
	}

	if len(input.Defaults) > 0 {
		prompt.Default = input.Defaults[0]
	}

	if input.Help != "" {
		prompt.Help = input.Help
	}

	question := &survey.Question{
		Prompt: prompt,
	}

	if input.Validate != nil {
		question.Validate = survey.Validator(input.Validate)
	}

	if input.Transform != nil {
		question.Transform = survey.Transformer(input.Transform)
	}

	return selected, survey.Ask([]*survey.Question{question}, &selected, surveyOpts...)
}

func (x *Survey) SelectMulti(input *Select) ([]string, error) {
	var selected []string

	if len(input.Options) == 0 {
		return selected, nil
	}

	surveyOpts := []survey.AskOpt{
		survey.WithStdio(x.In, x.Out, x.Err),
	}

	options := make([]string, 0, len(input.Options))
	for _, opt := range input.Options {
		if v, ok := opt.(string); ok {
			options = append(options, v)
		}
	}

	prompt := &survey.MultiSelect{
		Message: input.Message,
		Options: options,
	}

	if len(input.Defaults) > 0 {
		prompt.Default = input.Defaults
	}

	if input.Help != "" {
		prompt.Help = input.Help
	}

	question := &survey.Question{
		Prompt: prompt,
	}

	if input.Validate != nil {
		question.Validate = survey.Validator(input.Validate)
	}

	if input.Transform != nil {
		question.Transform = survey.Transformer(input.Transform)
	}

	return selected, survey.Ask([]*survey.Question{question}, &selected, surveyOpts...)
}

// TransformOnlyId is a `Transformer`. It receives an answer value and returns
// a copy of the answer with only the first word, which expected to be the ID of
// the resource.
func TransformOnlyId(answer interface{}) interface{} {
	switch ans := answer.(type) {

	// Select.
	case core.OptionAnswer:
		return core.OptionAnswer{
			Index: ans.Index,
			Value: strings.Fields(ans.Value)[0],
		}

	// Multi-select.
	case []core.OptionAnswer:
		out := make([]core.OptionAnswer, len(ans))
		for i, o := range ans {
			out[i] = core.OptionAnswer{
				Index: o.Index,
				Value: strings.Fields(o.Value)[0],
			}
		}
		return out
	}

	return answer
}

func ValidateInt64(answer interface{}) error {
	v, ok := answer.(string)
	if !ok {
		return fmt.Errorf("survey: unsupported answer type: %T", answer)
	}

	if _, err := strconv.ParseInt(v, 10, 64); err != nil {
		return fmt.Errorf("survey: %v", err)
	}

	return nil
}
