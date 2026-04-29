package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRootCommandShowsHelpWithoutSubcommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := newRootCommand(&stdout, &stderr)
	command.SetArgs(nil)

	require.NoError(t, command.Execute())
	text := stdout.String() + stderr.String()
	require.Contains(t, text, "console")
	require.Contains(t, text, "web")
	require.Contains(t, text, "sessions")
}

func TestNewRootCommandRegistersExplicitCommands(t *testing.T) {
	t.Parallel()

	command := newRootCommand(&bytes.Buffer{}, &bytes.Buffer{})

	found, _, err := command.Find([]string{"console"})
	require.NoError(t, err)
	require.Equal(t, "console", found.Name())

	found, _, err = command.Find([]string{"web"})
	require.NoError(t, err)
	require.Equal(t, "web", found.Name())

	found, _, err = command.Find([]string{"sessions", "list"})
	require.NoError(t, err)
	require.Equal(t, "list", found.Name())

	found, _, err = command.Find([]string{"sessions", "inspect"})
	require.NoError(t, err)
	require.Equal(t, "inspect", found.Name())

	found, _, err = command.Find([]string{"sessions", "delete"})
	require.NoError(t, err)
	require.Equal(t, "delete", found.Name())
}
