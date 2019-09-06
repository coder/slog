package console

import (
	"strconv"
	"strings"

	"go.coder.com/slog/slogcore"
)

// consoleMarshaller marshals a fieldValue into a human readable format.
type consoleMarshaller struct {
	out       strings.Builder
	indentstr string
}

func Fields(m slogcore.Map) string {
	var y consoleMarshaller
	y.marshal(m)
	return y.out.String()
}

func (y *consoleMarshaller) line() {
	y.out.WriteString("\n" + y.indentstr)
}

func (y *consoleMarshaller) s(s string) {
	y.out.WriteString(s)
}

func (y *consoleMarshaller) indent() {
	// 2 works really well for yaml because of arrays.
	// E.g. if you have a map inside of an list it looks like:
	//
	// - field: val
	//   field: val
	//
	// See how the second field automatically gets indented to the correct level?
	// We can use \t later and special case but for now this is simpler.
	y.indentstr += "  "
}

func (y *consoleMarshaller) unindent() {
	y.indentstr = y.indentstr[:len(y.indentstr)-2]
}

func (y *consoleMarshaller) marshal(v slogcore.Value) {
	switch v := v.(type) {
	case slogcore.String:
		// Ensures indentation.
		y.indent()
		// Replaces every newline with a newline plus the correct indentation.
		y.s(strings.ReplaceAll(string(v), "\n", "\n"+y.indentstr))
		y.unindent()
	case slogcore.Bool:
		y.s(strconv.FormatBool(bool(v)))
	case slogcore.Float:
		y.s(strconv.FormatFloat(float64(v), 'f', -1, 64))
	case slogcore.Int:
		y.s(strconv.FormatInt(int64(v), 10))
	case slogcore.Uint:
		y.s(strconv.FormatUint(uint64(v), 10))
	case slogcore.Map:
		for i, f := range v {
			if i > 0 {
				// Add newline before every field except first.
				y.line()
			}

			y.s(quote(f.Name) + ":")

			y.marshalSub(f.Value, true)
		}
	case slogcore.List:
		y.indent()
		for _, v := range v {
			y.line()
			y.s("-")

			if _, ok := v.(slogcore.List); !ok {
				// Non list values begin with the -.
				y.s(" ")
			}
			y.marshalSub(v, false)
		}
		y.unindent()
	case nil:
		y.s("null")
	default:
		panicf("unknown fieldValue kind of type %T and value %#v", y, y)
	}
}

func (y *consoleMarshaller) marshalSub(v slogcore.Value, isParentMap bool) {
	switch v := v.(type) {
	case slogcore.Map:
		if isParentMap && len(v) == 0 {
			// Nothing to output for this field. Without this line, we get additional newlines due to below code as
			// the map field is expected to start on the next line given it is in a parentMap.
			// See the emptyStruct test.
			return
		}

		y.indent()

		if isParentMap {
			// If we are not inside a list and the value is a map, we need a newline.
			// In other words, if we are inside a list and the value is a map, we want to start
			// it with the `-` of the list.
			y.line()
		}
	case slogcore.List:
	default:
		if isParentMap {
			// Non map and non list values in structs begin on the same line with a space between the key and value.
			y.s(" ")
		}
	}

	y.marshal(v)

	if _, ok := v.(slogcore.Map); ok {
		y.unindent()
	}
}

// quotes quotes a string so that it is suitable
// as a key for a map or in general some output that
// cannot span multiple lines or have weird characters.
func quote(key string) string {
	// strconv.Quote does not quote an empty string so we need this.
	if key == "" {
		return `""`
	}

	// Replace spaces in the map keys with underscores.
	key = strings.ReplaceAll(key, " ", "_")

	quoted := strconv.Quote(key)
	// If the key doesn't need to be quoted, don't quote it.
	// We do not use strconv.CanBackquote because it doesn't
	// account tabs.
	if quoted[1:len(quoted)-1] == key {
		return key
	}
	return quoted
}
