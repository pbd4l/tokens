package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	cmd := generateCmd()
	err := cmd.Flags().Set("number", "5")
	require.NoError(t, err)
	err = cmd.Flags().Set("seed", "1")
	require.NoError(t, err)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	b, err := ioutil.ReadAll(&stdout)
	require.NoError(t, err)
	require.Equal(t, `xvlbzgb
aicmraj
wwhthct
cuaxhxk
qfdafpl
`, string(b))
}
