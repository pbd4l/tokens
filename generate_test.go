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

	err = Generate([]string{
		"-file", f.Name(),
		"-n", "5",
		"-seed", "1",
	})
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
