package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	f, err := ioutil.TempFile("", "tokens.txt")
	require.Nil(t, err)
	defer os.Remove(f.Name())

	cmd := generateCmd()
	err = cmd.Flags().Set("file", f.Name())
	require.Nil(t, err)
	err = cmd.Flags().Set("number", "5")
	require.Nil(t, err)
	err = cmd.Flags().Set("seed", "1")
	require.Nil(t, err)

	err = cmd.Execute()
	require.Nil(t, err)

	b, err := ioutil.ReadAll(f)
	require.Nil(t, err)
	require.Equal(t, `xvlbzgb
aicmraj
wwhthct
cuaxhxk
qfdafpl
`, string(b))
}
