package parser

import (
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseString2CPool(t *testing.T) {
	const expectedValue = "com.example.filter.this.is.a.very.long.setting.value.that.exceeds." +
		"one.hundred.twenty.eight.characters.to.ensure.immediate.string.pool.encoding" +
		".two.in.jfr.format.padding.padding.padding"

	f, err := os.Open("testdata/string2.jfr.gz")
	require.NoError(t, err)
	defer f.Close()

	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gz.Close()

	data, err := io.ReadAll(gz)
	require.NoError(t, err)

	p := NewParser(data, Options{})

	cpoolChecked := false
	var foundValue string
	for {
		typ, err := p.ParseEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}

		if !cpoolChecked {
			assert.NotEmpty(t, p.TypeMap.CPoolStrings, "CPoolStrings should be populated after first chunk is parsed")
			cpoolChecked = true
		}

		if typ == p.TypeMap.T_ACTIVE_SETTING && p.ActiveSetting.Name == "customFilter" {
			foundValue = p.ActiveSetting.Value
		}
	}

	assert.True(t, cpoolChecked, "should have parsed at least one event")
	assert.Equal(t, expectedValue, foundValue,
		"ActiveSetting customFilter value should be resolved from constant pool via encoding 2")
}
