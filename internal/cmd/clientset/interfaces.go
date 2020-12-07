package clientset

import (
	"errors"

	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/editor"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/survey"
	"github.com/spotinst/spotctl/internal/thirdparty"
	"github.com/spotinst/spotctl/internal/writer"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("clients: not implemented")

type (
	// Factory interface represents a clients factory that creates instances of
	// each client type. For example, to create an instance of the cloud provider
	// client interface, call the following method Clientset.NewCloud().
	Factory interface {
		// NewSpotClient returns an instance of spot.Client.
		NewSpotClient(options ...spot.ClientOption) (spot.Client, error)

		// NewCloud returns an instance of cloud.Interface.
		NewCloud(name cloud.ProviderName, options ...cloud.ProviderOption) (cloud.Interface, error)

		// NewCommand returns an instance of thirdparty.Command.
		NewCommand(name thirdparty.CommandName) (thirdparty.Command, error)

		// NewSurvey returns an instance of survey.Interface.
		NewSurvey() (survey.Interface, error)

		// NewDepManager returns an instance of dep.Manager.
		NewDepManager() (dep.Manager, error)

		// NewEditor returns an instance of editor.Editor.
		NewEditor() (editor.Editor, error)

		// NewWriter returns an instance of writer.Writer.
		NewWriter(format writer.Format) (writer.Writer, error)
	}
)
