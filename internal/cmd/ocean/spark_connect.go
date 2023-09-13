package ocean

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	spotctlerrors "github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
	"os"
)

type SocketServer struct {
	conn *websocket.Conn
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
	return x.opts.Validate()
}

func (x *CmdSparkConnect) run(ctx context.Context) error {
	log.Infof("Spark connect will now run")
	_, err := x.NewWebSocketServer()
	if err != nil {
		log.Errorf("could not connect to websocket server %w", err)
		return err
	}

	return nil
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
	fs.StringVar(&x.WsUrl, flags.FlagOFASWsUrl, x.AppID, "web socket url")
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
		errg.Add(spotctlerrors.Required(flags.FlagOFASWsUrl))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}

func (x *CmdSparkConnect) NewWebSocketServer() (*SocketServer, error) {

	// todo Is there a some other way to get the env params from options?  Perhaps just use cmd options
	token := os.Getenv("SPOTINST_TOKEN")
	account := os.Getenv("SPOTINST_ACCOUNT")

	clusterID := x.opts.ClusterID
	appID := x.opts.AppID
	baseURL := x.opts.WsUrl

	address := fmt.Sprintf("%s/ocean/spark/cluster/%s/app/%s/connect?accountId=%s", baseURL, clusterID, appID, account)
	log.Infof("Starting websocket server on address %s", address)

	header := http.Header{"Authorization": []string{"Bearer " + token}}
	conn, resp, err := websocket.DefaultDialer.Dial(address, header)

	if err != nil {
		if err == websocket.ErrBadHandshake {
			log.Errorf("handshake failed with status %d", resp.StatusCode)
		}
		return nil, err
	}
	startSocketServer(conn)
	return &SocketServer{conn: conn}, nil
}

func startSocketServer(wsConn *websocket.Conn) {
	ln, err := net.Listen("tcp", ":15002")
	if err != nil {
		log.Errorf("handshake failed with status %w", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("Accept error: %w", err)
			return
		}

		go func() {
			err := handleConnection(conn, wsConn)
			if err != nil {
				log.Errorf("handle connection error: %w", err)
			}
		}()
	}
}

func handleConnection(conn net.Conn, wsConn *websocket.Conn) error {
	defer conn.Close()

	g, _ := errgroup.WithContext(context.Background())

	// Websocket to Upstream
	g.Go(func() error {
		return toUpstream(conn, wsConn)
	})

	// Upstream to websocket
	g.Go(func() error {
		return fromUpstream(conn, wsConn)
	})

	return g.Wait()
}

func fromUpstream(upstream io.Reader, downstream *websocket.Conn) error {
	for {
		buf := make([]byte, 1024)
		readFromUpstream, err := upstream.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if errors.Is(err, io.EOF) {
			log.Debugf("Upstream closed")
			break
		}

		//log.Debugf("Got %d bytes from upstream, will write back to peer", readFromUpstream)
		err = downstream.WriteMessage(websocket.BinaryMessage, buf[:readFromUpstream]) //wsutil.WriteServerBinary(downstream, buf[:readFromUpstream])
		if err != nil {
			return err
		}
	}

	return nil
}

func toUpstream(downstream io.Writer, upstream *websocket.Conn) error {
	for {
		_, msg, err := upstream.ReadMessage() //wsutil.ReadClientBinary(downstream)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if errors.Is(err, io.EOF) {
			log.Infof("Client closed")
			break
		}

		// log.Debugf("Read %d bytes from peer", len(msg))

		_, err = downstream.Write(msg)
		if err != nil {
			return err
		}
		//log.Debugf("Wrote %d bytes to upstream connect API", wroteUpstream)
	}

	return nil
}
