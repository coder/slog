package core

import (
	"sort"
	"strconv"
	"strings"

	structpb "github.com/golang/protobuf/ptypes/struct"
)

func quote(s string) string {
	if s == "" {
		return `""`
	}
	quoted := strconv.Quote(s)
	if quoted[1:len(quoted)-1] != s {
		return quoted
	}
	return s
}

type yamlMarshaller struct {
	out       strings.Builder
	indentstr string
}

func marshalFields(f map[string]*structpb.Value) string {
	var y yamlMarshaller
	sv := pbstruct(&structpb.Struct{
		Fields: f,
	})
	y.marshal(sv)
	return y.out.String()
}

func (y *yamlMarshaller) line() {
	y.out.WriteString("\n" + y.indentstr)
}

func (y *yamlMarshaller) s(s string) {
	y.out.WriteString(s)
}

func (y *yamlMarshaller) indent() {
	// 2 works really well for yaml because of arrays.
	// E.g. if you have a struct inside of an list it looks like:
	//
	// - field: val
	//   field: val
	//
	// See how the second field automatically gets indented to the correct level?
	// We can use \t later and special case but for now this is simpler.
	y.indentstr += "  "
}

func (y *yamlMarshaller) unindent() {
	y.indentstr = y.indentstr[:len(y.indentstr)-2]
}

// In the future, its worth revisiting this code to see if the logic can be simplified.
func (y *yamlMarshaller) marshal(v *structpb.Value) {
	switch v := v.Kind.(type) {
	case *structpb.Value_StringValue:
		y.s(strings.NewReplacer("\n", "\n"+y.indentstr+"\t").Replace(v.StringValue))
	case *structpb.Value_BoolValue:
		y.s(strconv.FormatBool(v.BoolValue))
	case *structpb.Value_NumberValue:
		y.s(strconv.FormatFloat(v.NumberValue, 'f', -1, 64))
	case *structpb.Value_StructValue:
		var keys []string
		for k := range v.StructValue.Fields {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for i, k := range keys {
			if i > 0 {
				// Add newline before every field except first.
				y.line()
			}

			fv := v.StructValue.Fields[k]

			k = strings.NewReplacer(" ", "_").Replace(k)
			y.s(quote(k) + ":")

			isStruct := fv.GetStructValue() != nil
			if isStruct {
				// Nothing to output for this field. Without this line, we get additional newlines due to below code as
				// the struct field is expected to start on the next line. Try outputting map[string]struct{} without this
				// or looking at the map[string]struct{} test.
				if len(fv.GetStructValue().Fields) == 0 {
					continue
				}
			} else if fv.GetListValue() == nil {
				// Non struct and non list values begin on the same line.
				y.s(" ")
			}

			y.marshalSub(fv, false)

		}
	case *structpb.Value_ListValue:
		y.indent()
		for _, v := range v.ListValue.Values {
			y.line()
			y.s("-")

			if v.GetListValue() == nil {
				// Non list values begin with the "-".
				y.s(" ")
			}
			y.marshalSub(v, true)
		}
		y.unindent()
	case *structpb.Value_NullValue:
		y.s("nil")
	default:
		panicf("unknown structpb kind of type %T and value %#v", y, y)
	}
}

func (y *yamlMarshaller) marshalSub(fv *structpb.Value, list bool) {
	isStruct := fv.GetStructValue() != nil
	if isStruct {
		y.indent()
	}
	// If we are inside a list and the value is a struct, we want to start it with the `-` of the list.
	if !list && isStruct {
		y.line()
	}
	y.marshal(fv)
	if isStruct {
		y.unindent()
	}
}
