package config

import (
	"context"
	"testing"

	_ "github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/obscure"
	"github.com/rclone/rclone/fs/rc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testName = "configTestNameForRc"

func TestRc(t *testing.T) {
	// Create the test remote
	call := rc.Calls.Get("config/create")
	assert.NotNil(t, call)
	in := rc.Params{
		"name": testName,
		"type": "local",
		"parameters": rc.Params{
			"test_key": "sausage",
		},
	}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.Nil(t, out)
	assert.Equal(t, "local", FileGet(testName, "type"))
	assert.Equal(t, "sausage", FileGet(testName, "test_key"))

	// The sub tests rely on the remote created above but they can
	// all be run independently

	t.Run("Dump", func(t *testing.T) {
		call := rc.Calls.Get("config/dump")
		assert.NotNil(t, call)
		in := rc.Params{}
		out, err := call.Fn(context.Background(), in)
		require.NoError(t, err)
		require.NotNil(t, out)

		require.NotNil(t, out[testName])
		config := out[testName].(rc.Params)

		assert.Equal(t, "local", config["type"])
		assert.Equal(t, "sausage", config["test_key"])
	})

	t.Run("Get", func(t *testing.T) {
		call := rc.Calls.Get("config/get")
		assert.NotNil(t, call)
		in := rc.Params{
			"name": testName,
		}
		out, err := call.Fn(context.Background(), in)
		require.NoError(t, err)
		require.NotNil(t, out)

		assert.Equal(t, "local", out["type"])
		assert.Equal(t, "sausage", out["test_key"])
	})

	t.Run("ListRemotes", func(t *testing.T) {
		call := rc.Calls.Get("config/listremotes")
		assert.NotNil(t, call)
		in := rc.Params{}
		out, err := call.Fn(context.Background(), in)
		require.NoError(t, err)
		require.NotNil(t, out)

		var remotes []string
		err = out.GetStruct("remotes", &remotes)
		require.NoError(t, err)

		assert.Contains(t, remotes, testName)
	})

	t.Run("Update", func(t *testing.T) {
		call := rc.Calls.Get("config/update")
		assert.NotNil(t, call)
		in := rc.Params{
			"name": testName,
			"parameters": rc.Params{
				"test_key":  "rutabaga",
				"test_key2": "cabbage",
			},
		}
		out, err := call.Fn(context.Background(), in)
		require.NoError(t, err)
		assert.Nil(t, out)

		assert.Equal(t, "local", FileGet(testName, "type"))
		assert.Equal(t, "rutabaga", FileGet(testName, "test_key"))
		assert.Equal(t, "cabbage", FileGet(testName, "test_key2"))
	})

	t.Run("Password", func(t *testing.T) {
		call := rc.Calls.Get("config/password")
		assert.NotNil(t, call)
		in := rc.Params{
			"name": testName,
			"parameters": rc.Params{
				"test_key":  "rutabaga",
				"test_key2": "cabbage",
			},
		}
		out, err := call.Fn(context.Background(), in)
		require.NoError(t, err)
		assert.Nil(t, out)

		assert.Equal(t, "local", FileGet(testName, "type"))
		assert.Equal(t, "rutabaga", obscure.MustReveal(FileGet(testName, "test_key")))
		assert.Equal(t, "cabbage", obscure.MustReveal(FileGet(testName, "test_key2")))
	})

	// Delete the test remote
	call = rc.Calls.Get("config/delete")
	assert.NotNil(t, call)
	in = rc.Params{
		"name": testName,
	}
	out, err = call.Fn(context.Background(), in)
	require.NoError(t, err)
	assert.Nil(t, out)
	assert.Equal(t, "", FileGet(testName, "type"))
	assert.Equal(t, "", FileGet(testName, "test_key"))
}

func TestRcProviders(t *testing.T) {
	call := rc.Calls.Get("config/providers")
	assert.NotNil(t, call)
	in := rc.Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	var registry []*fs.RegInfo
	err = out.GetStruct("providers", &registry)
	require.NoError(t, err)
	foundLocal := false
	for _, provider := range registry {
		if provider.Name == "local" {
			foundLocal = true
			break
		}
	}
	assert.True(t, foundLocal, "didn't find local provider")
}
