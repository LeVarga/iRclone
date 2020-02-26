// Sync files and directories to and from local and remote object stores
//
// Nick Craig-Wood <nick@craig-wood.com>
package rclone

import (
	"bytes"
	"context"
	_ "github.com/rclone/rclone/backend/all" // import all backends
	"github.com/rclone/rclone/lib/oauthutil"

	//_ "github.com/rclone/rclone/cmd/all"     // import all commands
	"github.com/rclone/rclone/cmd/serve/httplib"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/cmd/rc"
	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/fs/rc/rcserver"
	_ "github.com/rclone/rclone/lib/plugin" // import plugins
	"time"
	_"github.com/rclone/rclone/fs/sync"
)

var server *rcserver.Server

func StopRC() {
	server.Close()
}

func StartRC() error {
	httpOptions := httplib.DefaultOpt
	httpOptions.ListenAddr = "localhost:5572"
	var error error
	server, error = rcserver.Start(&rc.Options{
		HTTPOptions:              httpOptions,
		Enabled:                  true,
		Serve:                    true,
		Files:                    "",
		NoAuth:                   true,
		WebUI:                    false,
		WebGUIUpdate:             false,
		WebGUIFetchURL:           "",
		AccessControlAllowOrigin: "",
		JobExpireDuration:        24 * time.Hour,
		JobExpireInterval:        time.Minute,
	})
	if error != nil {
		return error
	}
	return nil
}

type Operation struct {
	args []string
	outBuf bytes.Buffer
}

func (o *Operation) AddArgs(arg string) {
	o.args = append(o.args, arg)
}

func (o *Operation) Run() (string, error) {
	return rcclient.Run(context.Background(), o.args)
}

func SetConfigPath(configPath string) {
	config.ConfigPath = configPath
}

func GetAuthState() string {
	return oauthutil.AuthState
}