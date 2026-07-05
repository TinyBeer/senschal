package envmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// parseVersion
// ---------------------------------------------------------------------------

func TestParseVersion_Simple(t *testing.T) {
	v, err := parseVersion("24.0.5")
	require.NoError(t, err)
	assert.Equal(t, []int{24, 0, 5}, v)
}

func TestParseVersion_Single(t *testing.T) {
	v, err := parseVersion("1")
	require.NoError(t, err)
	assert.Equal(t, []int{1}, v)
}

func TestParseVersion_Empty(t *testing.T) {
	_, err := parseVersion("")
	assert.Error(t, err)
}

func TestParseVersion_Invalid(t *testing.T) {
	_, err := parseVersion("24.a.5")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// compareVersion - exact match
// ---------------------------------------------------------------------------

func TestCompareVersion_ExactMatch(t *testing.T) {
	ok, err := compareVersion("24.0.5", "24.0.5")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCompareVersion_ExactMismatch(t *testing.T) {
	ok, err := compareVersion("24.0.5", "24.0.6")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareVersion_ExactDifferentLength(t *testing.T) {
	ok, err := compareVersion("24.0", "24.0.5")
	require.NoError(t, err)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// compareVersion - prefix match (^)
// ---------------------------------------------------------------------------

func TestCompareVersion_PrefixMatch(t *testing.T) {
	ok, err := compareVersion("^24.0", "24.0.5")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCompareVersion_PrefixMatch_MinorGreater(t *testing.T) {
	ok, err := compareVersion("^24.0", "24.1.0")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCompareVersion_PrefixMatch_MinorLess(t *testing.T) {
	ok, err := compareVersion("^24.1", "24.0.5")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareVersion_PrefixMatch_MajorLess(t *testing.T) {
	ok, err := compareVersion("^25.0", "24.0.5")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareVersion_PrefixMatch_ExpectedLonger(t *testing.T) {
	// expected version has more segments than actual -> error
	_, err := compareVersion("^24.0.5", "24.0")
	assert.Error(t, err)
}

func TestCompareVersion_PrefixMatch_Exact(t *testing.T) {
	ok, err := compareVersion("^24.0", "24.0")
	require.NoError(t, err)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// compareVersion - invalid inputs
// ---------------------------------------------------------------------------

func TestCompareVersion_InvalidExpected(t *testing.T) {
	_, err := compareVersion("24.a", "24.0")
	assert.Error(t, err)
}

func TestCompareVersion_InvalidActual(t *testing.T) {
	_, err := compareVersion("24.0", "24.b")
	assert.Error(t, err)
}
