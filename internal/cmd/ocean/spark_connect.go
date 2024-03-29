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
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	defaultWsUrl            = "wss://api.spotinst.io"
	defaultSparkConnectPort = "15002"
	defaultHivePort         = "10000"
	defaultEndpoint         = "connect"
)

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
		EndPoint  string
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

	socketConnection, err := x.connectToSocketServer(ctx)
	if err != nil {
		log.Errorf("could not connect to websocket server %w", err)
		return err
	}

	log.Infof("Starting websocket server on address %s", socketConnection.RemoteAddr().String())

	// Start ping ticker to keep connection alive
	pingTicker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := socketConnection.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Errorf("Failed to send ping: %v", err)
				}
			}
		}
	}()
	defer pingTicker.Stop()

	ln, err := net.Listen("tcp", ":"+x.opts.Port)
	if err != nil {
		log.Errorf("handshake failed with status %w", err)
		return err
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			log.Errorf("error closing listener %w", err)
		}
	}(ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("Accept error: %w", err)
			return err
		}

		go func() {
			err := handleSocketConnection(ctx, conn, socketConnection)
			if err != nil {
				log.Errorf("handle connection error: %w", err)
			}
		}()
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
	fs.StringVar(&x.WsUrl, flags.FlagOFASWsUrl, x.WsUrl, "web socket url. Default is wss://api.spotinst.io")
	fs.StringVar(&x.Port, "port", x.Port, "spark connection port. Default is 15002")
	fs.StringVar(&x.EndPoint, "endpoint", x.EndPoint, "spark connection endpoint. Default is connect, hive is also supported")
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

	if x.EndPoint == "" {
		x.EndPoint = defaultEndpoint
	} else if x.EndPoint != defaultEndpoint && x.EndPoint != "hive" {
		errg.Add(fmt.Errorf("endpoint %s is not supported", x.EndPoint))
	}

	if x.Port == "" {
		if x.EndPoint == defaultEndpoint {
			x.Port = defaultSparkConnectPort
		} else {
			x.Port = defaultHivePort
		}
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}

func (x *CmdSparkConnect) connectToSocketServer(ctx context.Context) (*websocket.Conn, error) {
	cfg := spotinst.DefaultConfig()
	cred, err := cfg.Credentials.Get()
	if err != nil {
		return nil, err
	}

	clusterID := x.opts.ClusterID
	appID := x.opts.AppID
	baseURL := x.opts.WsUrl
	endPoint := x.opts.EndPoint

	address := fmt.Sprintf("%s/ocean/spark/cluster/%s/app/%s/%s?accountId=%s", baseURL, clusterID, appID, endPoint, cred.Account)
	log.Infof("Starting websocket server on address %s", address)

	header := http.Header{"Authorization": []string{"Bearer " + cred.Token}}
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, address, header)

	if err != nil {
		if errors.Is(err, websocket.ErrBadHandshake) {
			log.Errorf("handshake failed with status %d", resp.StatusCode)
		}
		return nil, err
	}

	return conn, nil
}

func handleSocketConnection(ctx context.Context, conn net.Conn, wsConn *websocket.Conn) error {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("error closing connection %w", err)
		}
	}(conn)

	g, _ := errgroup.WithContext(ctx)

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

		err = downstream.WriteMessage(websocket.BinaryMessage, buf[:readFromUpstream])
		if err != nil {
			return err
		}
	}

	return nil
}

func toUpstream(downstream io.Writer, upstream *websocket.Conn) error {
	for {
		_, msg, err := upstream.ReadMessage()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if errors.Is(err, io.EOF) {
			log.Infof("Client closed")
			break
		}

		_, err = downstream.Write(msg)
		if err != nil {
			return err
		}
	}

	return nil
}
