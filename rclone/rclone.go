// Sync files and directories to and from local and remote object stores
//
// Nick Craig-Wood <nick@craig-wood.com>
package rclone

import (
	"context"
	_ "github.com/rclone/rclone/backend/all" // import all backends
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/fs/rc/rcserver"
	_ "github.com/rclone/rclone/fs/sync"
	"github.com/rclone/rclone/lib/oauthutil"
	_ "github.com/rclone/rclone/lib/plugin" // import plugins
	"time"
)

var server *rcserver.Server

func StopRC() {
	server.Shutdown()
}

func StartRC() error {
	httpOptions := rc.DefaultOpt.HTTP
	httpOptions.ListenAddr = []string{"localhost:5572"}
	var error error
	server, error = rcserver.Start(context.Background(), &rc.Options{
		HTTP:              httpOptions,
		Enabled:           true,
		Serve:             true,
		Files:             "",
		NoAuth:            true,
		WebUI:             false,
		WebGUIUpdate:      false,
		WebGUIFetchURL:    "",
		JobExpireDuration: 24 * time.Hour,
		JobExpireInterval: time.Minute,
	})
	if error != nil {
		return error
	}
	return nil
}

func SetConfigPath(configPath string) {
	config.SetConfigPath(configPath)
	configfile.Install()
}

func GetAuthState() string {
	return oauthutil.AuthState
}
