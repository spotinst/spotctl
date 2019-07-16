package clients

import (
	"errors"

	"github.com/spotinst/spotinst-cli/internal/cloud"
	"github.com/spotinst/spotinst-cli/internal/dep"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/survey"
	"github.com/spotinst/spotinst-cli/internal/thirdparty"
	"github.com/spotinst/spotinst-cli/internal/writer"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("clients: not implemented")

type (
	// Factory interface represents a clients factory that creates instances of
	// each client type. For example, to create an instance of the cloud provider
	// client interface, call the following method Clients.NewCloud().
	Factory interface {
		// NewSpotinst returns an instance of Spotinst interface.
		NewSpotinst(options ...spotinst.ClientOption) (spotinst.Interface, error)

		// NewCloud returns an instance of cloud provider by name.
		NewCloud(name cloud.ProviderName) (cloud.Interface, error)

		// NewCommand returns an instance of a third-party command by name.
		NewCommand(name thirdparty.CommandName) (thirdparty.Command, error)

		// NewSurvey returns an instance of survey interface.
		NewSurvey() (survey.Interface, error)

		// NewDep returns an instance of dependency manager interface.
		NewDep() (dep.Interface, error)

		// NewWriter returns an instance of writer interface.
		NewWriter(format writer.Format) (writer.Writer, error)
	}
)
