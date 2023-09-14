package ocean

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	spotctlerrors "github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
)

const (
	defaultWsUrl            = "wss://api.spotinst.io"
	defaultSparkConnectPort = "15002"
)

type wsConn struct {
	net.Conn
}

type (
	CmdSparkConnect struct {
		cmd  *cobra.Command
		opts CmdSparkConnectOptions
	}

	CmdSparkConnectOptions struct {
		*CmdSparkOptions
		WsUrl     string
		ClusterID string
		AppID     string
		Port      string
	}
)

func NewCmdSparkConnect(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkConnect(opts).cmd
}

func newCmdSparkConnect(opts *CmdSparkOptions) *CmdSparkConnect {
	var cmd CmdSparkConnect

	cmd.cmd = &cobra.Command{
		Use:           "connect",
		Short:         "connect to Ocean Spark",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"t"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdSparkConnect) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return nil
}

func (x *CmdSparkConnect) Run(ctx context.Context) error {
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

func (x *CmdSparkConnect) survey(_ context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkConnect) log(_ context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkConnect) validate(_ context.Context) error {
	// perhaps get credentials here
	return x.opts.Validate()
}

func (x *CmdSparkConnectOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkConnectOptions) initDefaults(opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
}

func (x *CmdSparkConnectOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOFASClusterID, x.ClusterID, "id of the cluster")
	fs.StringVar(&x.AppID, flags.FlagOFASAppID, x.AppID, "id of the spark application")
	fs.StringVar(&x.WsUrl, flags.FlagOFASWsUrl, x.AppID, "web socket url. Default is wss://api.spotinst.io")
	fs.StringVar(&x.Port, "port", x.Port, "spark connection port. Default is 1502")
}

func (x *CmdSparkConnectOptions) Validate() error {
	errg := spotctlerrors.NewErrorGroup()

	if err := x.CmdSparkOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(spotctlerrors.Required(flags.FlagOFASClusterID))
	}

	if x.AppID == "" {
		errg.Add(spotctlerrors.Required(flags.FlagOFASAppID))
	}

	if x.WsUrl == "" {
		x.WsUrl = defaultWsUrl
	}

	if x.Port == "" {
		x.Port = defaultSparkConnectPort
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}

func (x *CmdSparkConnect) run(ctx context.Context) error {
	log.Infof("Spark connect will now run")

	return x.connect(ctx)

}

func (spkcon *CmdSparkConnect) connect(ctx context.Context) error {
	cfg := spotinst.DefaultConfig()
	cred, err := cfg.Credentials.Get()
	if err != nil {
		return err
	}

	token := cred.Token
	account := cred.Account
	clusterID := spkcon.opts.ClusterID
	appID := spkcon.opts.AppID
	baseURL := spkcon.opts.WsUrl

	// Setting up dialer with headers
	dialer := &ws.Dialer{
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
		},
	}

	address := fmt.Sprintf("%s/ocean/spark/cluster/%s/app/%s/connect?accountId=%s", baseURL, clusterID, appID, account)
	log.Infof("Connecting to WebSocket server on address %s", address)

	conn, _, _, err := dialer.Dial(ctx, address)
	if err != nil {
		return fmt.Errorf("unable to connect to spark connect API: %w", err)
	}
	log.Infof("Opened connection to spark connect API")
	defer closeConnection(conn)

	g, _ := errgroup.WithContext(ctx)

	// Read from WebSocket and write to Upstream
	g.Go(func() error {
		return toUpstream(conn, conn)
	})

	// Read from Upstream and write to WebSocket
	g.Go(func() error {
		return fromUpstream(conn, conn)
	})

	return g.Wait()
}

func (c *wsConn) Close() error {
	if err := wsutil.WriteServerMessage(c.Conn, ws.OpClose, nil); err != nil {
		log.Errorf("Unable to send close message")
	}

	log.Infof("Closing websocket connection")
	closeConnection(c.Conn)

	return nil
}

func closeConnection(conn net.Conn) {
	log.Infof("Closing connection to %s", conn.RemoteAddr())
	if err := conn.Close(); err != nil {
		log.Errorf("There was a problem closing connection to %s", conn.RemoteAddr())
	}
}

func websocketUpgrade(r *http.Request, w http.ResponseWriter) (net.Conn, error) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return nil, err
	}

	return &wsConn{
		Conn: conn,
	}, nil
}

func fromUpstream(upstream io.Reader, downstream io.Writer) error {
	for {
		buf := make([]byte, 1024)
		readFromUpstream, err := upstream.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if errors.Is(err, io.EOF) {
			log.Infof("Upstream closed")
			break
		}

		log.Infof("Got %d bytes from upstream, will write back to peer", readFromUpstream)
		err = wsutil.WriteServerBinary(downstream, buf[:readFromUpstream])
		if err != nil {
			return err
		}
	}

	return nil
}

func toUpstream(downstream io.ReadWriter, upstream io.Writer) error {
	for {
		msg, err := wsutil.ReadClientBinary(downstream)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if errors.Is(err, io.EOF) {
			log.Infof("Client closed")
			break
		}

		log.Debugf("Read %d bytes from peer", len(msg))

		wroteUpstream, err := upstream.Write(msg)
		if err != nil {
			return err
		}
		log.Debugf("Wrote %d bytes to upstream connect API", wroteUpstream)
	}

	return nil
}
