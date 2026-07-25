package theme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThemeFormatting(t *testing.T) {
	rawJSON := `{"name":"test","value":123}`
	pretty, err := PrettyFormatJSON([]byte(rawJSON))
	assert.NoError(t, err)
	assert.Contains(t, pretty, "\n  \"name\": \"test\"")

	dark := GetTheme("dark")
	styles := MakeStyles(dark)

	getStyle := styles.MethodStyle("GET")
	assert.NotNil(t, getStyle)

	statusStyle := styles.StatusStyle(200)
	assert.NotNil(t, statusStyle)
}
