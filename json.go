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

const (
	valueSep   = ", "
	null       = "null"
	startMap   = "{"
	endMap     = "}"
	startArray = "["
	endArray   = "]"
)

const (
	emptyMap   = startMap + endMap
	emptyArray = startArray + endArray
)

// Color used in human-readable terminal output, chosen to mimic the default `jq` colors.
var (
	stringColor = newStyle("#8ae234")
	boolColor   = newStyle("#34e2e2")
	numberColor = newStyle("#fce94f")
	nullColor   = newStyle("#ad7fa8")
)

// The default color profile determined by lipgloss, used by ResetColor.
var defaultProfile = lipgloss.ColorProfile()

// Colorize forces colorized output, even without a TTY.
func Colorize() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// ResetColor returns to lipgloss' default color handling.
func ResetColor() {
	lipgloss.SetColorProfile(defaultProfile)
}

// sprintf writes a string in the given color.
func sprintf(c *lipgloss.Style, format string, args ...interface{}) string {
	return c.Render(fmt.Sprintf(format, args...))
}

// marshal emits colorized JSON output. Adapted from TylerBrock/colorjson with instrument-specific modifications:
//   - Uses lipgloss rather than fatih/color.
//   - The caller determines the key color, for example based on log level.
//   - No newlines in the output.
//   - No indentation.
//   - No max string length.
//   - No sorting map keys.
//   - No input validation.
//   - Handles more built-in types.
func marshal(jsonObj Tags, keyColor *lipgloss.Style) []byte {
	buffer := bytes.Buffer{}
	marshalValue(jsonObj, &buffer, keyColor)

	return buffer.Bytes()
}

// marshalMap writes a JSON map.
func marshalMap(input map[string]interface{}, buf *bytes.Buffer, keyColor *lipgloss.Style) {
	remaining := len(input)

	if remaining == 0 {
		buf.WriteString(emptyMap)

		return
	}

	buf.WriteString(startMap)

	for key, val := range input {
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
func marshalArray(input []interface{}, buf *bytes.Buffer, keyColor *lipgloss.Style) {
	if len(input) == 0 {
		buf.WriteString(emptyArray)

		return
	}

	buf.WriteString(startArray)

	for i, v := range input {
		marshalValue(v, buf, keyColor)

		if i < len(input)-1 {
			buf.WriteString(valueSep)
		}
	}

	buf.WriteString(endArray)
}

// marshalValue handles a bunch of different built-in types.
//
//nolint:cyclop
func marshalValue(input interface{}, buf *bytes.Buffer, keyColor *lipgloss.Style) {
	switch val := input.(type) {
	case map[string]interface{}:
		marshalMap(val, buf, keyColor)
	case []interface{}:
		marshalArray(val, buf, keyColor)
	case string:
		marshalString(val, buf)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i := reflect.ValueOf(val).Int()
		buf.WriteString(sprintf(numberColor, "%d", i))
	case float32, float64:
		f := reflect.ValueOf(val).Float()
		buf.WriteString(sprintf(numberColor, strconv.FormatFloat(f, 'f', -1, 64)))
	case bool:
		buf.WriteString(sprintf(boolColor, (strconv.FormatBool(val))))
	case nil:
		buf.WriteString(sprintf(nullColor, null))
	case json.Number:
		buf.WriteString(sprintf(numberColor, val.String()))
	case error:
		buf.WriteString(sprintf(stringColor, "\"%s\"", val.Error()))
	case time.Time:
		buf.WriteString(sprintf(stringColor, "\"%s\"", val.UTC().Format(time.RFC3339)))
	case fmt.Stringer:
		buf.WriteString(sprintf(stringColor, "\"%s\"", val.String()))
	}
}

// marshalString writes a JSON string.
func marshalString(str string, buf *bytes.Buffer) {
	buf.WriteString(sprintf(stringColor, "\"%s\"", str))
}
