// Sync files and directories to and from local and remote object stores
//
// Nick Craig-Wood <nick@craig-wood.com>
package rclone

import (
	"context"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/fs/rc/rcserver"
	"github.com/rclone/rclone/lib/oauthutil"
	_ "github.com/rclone/rclone/backend/all"
	_ "github.com/rclone/rclone/cmd/all"
	_ "github.com/rclone/rclone/lib/plugin"
	"strings"
	"time"
)

var server *rcserver.Server

func StopRC() error {
	err := server.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

func StartRC(urls string) error {
	httpOptions := rc.DefaultOpt.HTTP
	httpOptions.ListenAddr = strings.Split(urls, ",")
	var err error = nil
	server, err = rcserver.Start(context.Background(), &rc.Options{
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
	if err != nil {
		return err
	}
	return nil
}

func SetConfigPath(configPath string) error {
	err := config.SetConfigPath(configPath)
	configfile.Install()
	if err != nil {
		return err
	}
	return nil
}

func GetAuthState() string {
	return oauthutil.AuthState
}
