package clientset

import (
	"errors"

	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/editor"
	"github.com/spotinst/spotctl/internal/spotinst"
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
		// NewSpotinst returns an instance of Spotinst interface.
		NewSpotinst(options ...spotinst.ClientOption) (spotinst.Interface, error)

		// NewCloud returns an instance of cloud provider by name.
		NewCloud(name cloud.ProviderName, options ...cloud.ProviderOption) (cloud.Interface, error)

		// NewCommand returns an instance of a third-party command by name.
		NewCommand(name thirdparty.CommandName) (thirdparty.Command, error)

		// NewSurvey returns an instance of survey interface.
		NewSurvey() (survey.Interface, error)

		// NewDep returns an instance of dependency manager interface.
		NewDep() (dep.Interface, error)

		// NewEditor returns an instance of an editor.
		NewEditor() (editor.Editor, error)

		// NewWriter returns an instance of writer interface.
		NewWriter(format writer.Format) (writer.Writer, error)
	}
)
