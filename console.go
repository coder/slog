package slog

import (
	"strconv"
	"strings"
)

// consoleMarshaller marshals a fieldValue into a human readable format.
type consoleMarshaller struct {
	out       strings.Builder
	indentstr string
}

func marshalFields(v fieldMap) string {
	var y consoleMarshaller
	y.marshal(v)
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

func (y *consoleMarshaller) marshal(v fieldValue) {
	switch v := v.(type) {
	case fieldString:
		// Ensures indentation.
		y.indent()
		// Replaces every newline with a newline plus the correct indentation.
		y.s(strings.ReplaceAll(string(v), "\n", "\n"+y.indentstr))
		y.unindent()
	case fieldBool:
		y.s(strconv.FormatBool(bool(v)))
	case fieldFloat:
		y.s(strconv.FormatFloat(float64(v), 'f', -1, 64))
	case fieldInt:
		y.s(strconv.FormatInt(int64(v), 10))
	case fieldUint:
		y.s(strconv.FormatUint(uint64(v), 10))
	case fieldMap:
		for i, f := range v {
			if i > 0 {
				// Add newline before every field except first.
				y.line()
			}

			y.s(quoteMapKey(f.name) + ":")

			y.marshalSub(f.value, true)
		}
	case fieldList:
		y.indent()
		for _, v := range v {
			y.line()
			y.s("-")

			if _, ok := v.(fieldList); !ok {
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

func (y *consoleMarshaller) marshalSub(v fieldValue, isParentMap bool) {
	switch v := v.(type) {
	case fieldMap:
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
	case fieldList:
	default:
		if isParentMap {
			// Non map and non list values in structs begin on the same line with a space between the key and value.
			y.s(" ")
		}
	}

	y.marshal(v)

	if _, ok := v.(fieldMap); ok {
		y.unindent()
	}
}

// quoteMapKey quotes a string so that it is suitable
// as a key for a map.
func quoteMapKey(key string) string {
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
