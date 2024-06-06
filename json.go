package instrument

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const valueSep = ", "
const null = "null"
const startMap = "{"
const endMap = "}"
const startArray = "["
const endArray = "]"

const emptyMap = startMap + endMap
const emptyArray = startArray + endArray

// Color used in human-readable terminal output, chosen to mimic the default jq colors.
var (
	stringColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ae234"))
	boolColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#34e2e2"))
	numberColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#fce94f"))
	nullColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ad7fa8"))
)

// The default color profile determined by lipgloss, used by ResetColor.
var defaultProfile = lipgloss.ColorProfile()

// Colorize forces colorized output. Beware of using this without a TTY.
func Colorize() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// ResetColor returns to lipgloss' default color handling.
func ResetColor() {
	lipgloss.SetColorProfile(defaultProfile)
}

// sprintf writes a string in the given color.
func sprintf(c lipgloss.Style, format string, args ...interface{}) string {
	return c.Render(fmt.Sprintf(format, args...))
}

// marshal emits colorized JSON output. Adapted from TylerBrock/colorjson with some instrument-specific modifications:
//   - Uses lipgloss instead of fatih/color.
//   - The caller determines the key color.
//   - No newlines in the output except for the end.
//   - No indentation.
//   - No string maximum length.
//   - No sorting map keys.
//   - No input validation.
//   - Handles additional built-in types.
func marshal(jsonObj interface{}, keyColor lipgloss.Style) ([]byte, error) {
	buffer := bytes.Buffer{}
	marshalValue(jsonObj, &buffer, keyColor)
	return buffer.Bytes(), nil
}

// marshalMap writes a JSON map.
func marshalMap(m map[string]interface{}, buf *bytes.Buffer, keyColor lipgloss.Style) {
	remaining := len(m)

	if remaining == 0 {
		buf.WriteString(emptyMap)
		return
	}

	buf.WriteString(startMap)

	for key, val := range m {
		buf.WriteString(sprintf(keyColor, "\"%s\": ", key))
		marshalValue(val, buf, keyColor)
		remaining--
		if remaining != 0 {
			buf.WriteString(valueSep)
		}
	}
	buf.WriteString(endMap)
}

// marshalArray writes a JSON array.
func marshalArray(a []interface{}, buf *bytes.Buffer, keyColor lipgloss.Style) {
	if len(a) == 0 {
		buf.WriteString(emptyArray)
		return
	}

	buf.WriteString(startArray)

	for i, v := range a {
		marshalValue(v, buf, keyColor)
		if i < len(a)-1 {
			buf.WriteString(valueSep)
		}
	}
	buf.WriteString(endArray)
}

// marshalValue handles a bunch of different built-in types.
func marshalValue(val interface{}, buf *bytes.Buffer, keyColor lipgloss.Style) {
	switch v := val.(type) {
	case Tags:
		marshalMap(v, buf, keyColor)
	case map[string]interface{}:
		marshalMap(v, buf, keyColor)
	case []interface{}:
		marshalArray(v, buf, keyColor)
	case string:
		marshalString(v, buf)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(v).Int()
		buf.WriteString(sprintf(numberColor, "%d", i))
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		buf.WriteString(sprintf(numberColor, strconv.FormatFloat(f, 'f', -1, 64)))
	case bool:
		buf.WriteString(sprintf(boolColor, (strconv.FormatBool(v))))
	case nil:
		buf.WriteString(sprintf(nullColor, null))
	case json.Number:
		buf.WriteString(sprintf(numberColor, v.String()))
	case error:
		buf.WriteString(sprintf(stringColor, "\"%s\"", v.Error()))
	case time.Time:
		buf.WriteString(sprintf(stringColor, "\"%s\"", v.UTC().Format(time.RFC3339)))
	case fmt.Stringer:
		buf.WriteString(sprintf(stringColor, "\"%s\"", v.String()))
	}
}

// marshalString writes a JSON string.
func marshalString(str string, buf *bytes.Buffer) {
	buf.WriteString(sprintf(stringColor, "\"%s\"", str))
}
