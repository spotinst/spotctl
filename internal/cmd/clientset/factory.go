package clientset

import (
	"io"

	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/editor"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/survey"
	"github.com/spotinst/spotctl/internal/thirdparty"
	"github.com/spotinst/spotctl/internal/writer"
)

type factory struct {
	in       io.Reader
	out, err io.Writer
}

func NewFactory(in io.Reader, out, err io.Writer) Factory {
	return &factory{
		in:  in,
		out: out,
		err: err,
	}
}

func (x *factory) NewSpotClient(options ...spot.ClientOption) (spot.Client, error) {
	log.Debugf("Instantiating new spot client")
	return spot.New(options...), nil
}

func (x *factory) NewCloud(name cloud.ProviderName, options ...cloud.ProviderOption) (cloud.Provider, error) {
	log.Debugf("Instantiating new cloud: %s", name)
	return cloud.GetInstance(name, options...)
}

func (x *factory) NewCommand(name thirdparty.CommandName) (thirdparty.Command, error) {
	log.Debugf("Instantiating new command: %s", name)
	return thirdparty.GetInstance(name, thirdparty.WithStdio(x.in, x.out, x.err))
}

func (x *factory) NewSurvey() (survey.Interface, error) {
	log.Debugf("Instantiating new survey")
	return survey.New(x.in, x.out, x.err), nil
}

func (x *factory) NewDepManager() (dep.Manager, error) {
	log.Debugf("Instantiating new dependency manager")
	return dep.NewManager(survey.New(x.in, x.out, x.err)), nil
}

func (x *factory) NewEditor() (editor.Editor, error) {
	log.Debugf("Instantiating new editor")
	return editor.New(x.in, x.out, x.err), nil
}

func (x *factory) NewWriter(format writer.Format) (writer.Writer, error) {
	log.Debugf("Instantiating new writer: %s", format)
	return writer.GetInstance(format, x.out)
}
