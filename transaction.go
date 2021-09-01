package fixedwidth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/charlieparkes/go-structs"
	"github.com/mitchellh/mapstructure"
)

const (
	SPACE rune = ' ' // UTF-8 32
)

type Field struct {
	Name    string
	Record  int
	Start   int
	End     int
	Literal string
}

type Transaction struct {
	Records    [][]byte
	Layout     Layout
	fieldCache map[string]string
	fieldNames map[string]pos
}

type pos struct {
	RecordIdx int
	FieldIdx  int
}

type Layout [][]Field

var layoutCache map[string]*Layout = map[string]*Layout{}

// NewTransaction allocates a new transaction and generates its layout using the target struct.
func NewTransaction(target interface{}) (*Transaction, error) {
	t := &Transaction{
		Records: make([][]byte, 0),
		Layout:  make([][]Field, 0),
	}

	// Cache layouts
	layoutName := structs.Name(target)
	if layout, ok := layoutCache[layoutName]; ok {
		t.Layout = *layout
		return t, nil
	}

	tags := structs.Tags(target, "txn")

	// Build layout from target struct
	for _, field := range structs.Fields(target) {
		f := Field{Name: field.Name}
		if val, ok := tags[field.Name]; ok {
			err := mapstructure.WeakDecode(val, &f)
			if err != nil {
				return nil, err
			}

			// Record is optional if len(txn)==1
			if f.Record == 0 {
				f.Record = 1
			}
			// End is optional if len(field)==1
			if f.End < f.Start {
				f.End = f.Start
			}

			for len(t.Layout) < f.Record {
				t.Layout = append(t.Layout, []Field{})
			}
			t.Layout[f.Record-1] = append(t.Layout[f.Record-1], f)
		}
	}

	layoutCache[layoutName] = &t.Layout

	return t, nil
}

// Append adds a record byte string to the transaction.
func (t *Transaction) Append(data ...[]byte) {
	for _, bs := range data {
		bsCopy := make([]byte, len(bs))
		copy(bsCopy, bs)
		t.Records = append(t.Records, bsCopy)
	}
}

// GetFields gets the value of known layout fields from the transaction records.
func (t *Transaction) GetFields() map[string]string {
	if t.fieldCache != nil {
		return t.fieldCache
	}

	l := 0
	for _, fields := range t.Layout {
		l += len(fields)
	}
	fields := make(map[string]string, l)

	for i, record := range t.Records {
		var layout []Field
		if len(t.Layout)-1 < i {
			continue // If user provided more records than we have layouts, skip them.
		} else {
			layout = t.Layout[i]
		}

		recordLen := len(record)
		for _, field := range layout {
			if field.Literal != "" {
				fields[field.Name] = field.Literal
			}

			// If line isn't long enough to even partially fill the field, skip the field.
			if recordLen < field.Start+1 {
				continue
			}

			// If line truncates a field, set high bound to max available.
			end := field.End + 1
			if recordLen < end {
				end = recordLen
			}

			fields[field.Name] = string(bytes.TrimSpace(record[field.Start:end]))
		}
	}

	return fields
}

// SetFields sets the values of known layout fields to the transaction records.
func (t *Transaction) SetFields(fields map[string]string) error {
	// Cache the position of each field in the layout so we can set them by name.
	if t.fieldCache == nil {
		t.fieldNames = map[string]pos{}
		for i, fields := range t.Layout {
			for j, f := range fields {
				t.fieldNames[f.Name] = pos{i, j}
			}
		}
	}

	if len(t.Records) == 0 {
		t.Records = make([][]byte, len(t.Layout))
	}

	// Write field data
	for name, val := range fields {
		var f Field
		if pos, ok := t.fieldNames[name]; !ok {
			return fmt.Errorf("tried to set field '%v' which does not exist in this layout", name)
		} else {
			f = t.Layout[pos.RecordIdx][pos.FieldIdx]
		}

		recordIdx := f.Record - 1

		// Initialize record
		if t.Records[recordIdx] == nil {
			t.Records[recordIdx] = []byte{}
		}

		// If the record is an empty bytestring, pad it up to where the field end.
		if len(t.Records[recordIdx]) < f.End+1 {
			t.Records[recordIdx] = PadRight(t.Records[recordIdx], f.End+1, SPACE)
		}

		// Write new field value to record
		if f.Literal != "" {
			val = f.Literal
		}
		highBound := f.End
		for offset, b := range []byte(val) {
			if i := f.Start + offset; i <= highBound {
				t.Records[recordIdx][i] = b
			}
		}
	}

	return nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.GetFields())
}

func (t *Transaction) Unmarshal(output interface{}) error {
	fields := t.GetFields()

	// Pick time format using output struct tags; otherwise, default to RFC3339.
	// TODO: switch to structs.FillStruct so we can get time layout for individual
	// fields, similar to structs.FillMap.
	dateLayout := time.RFC3339
	tags := structs.Tags(output, "txn")
	for _, fieldTags := range tags {
		if val, ok := fieldTags["time"]; ok {
			dateLayout = val
			break
		}
	}

	decodeHook := func(from reflect.Value, to reflect.Value) (interface{}, error) {
		toKind := to.Kind()
		if from.Kind() == reflect.String {
			fromStr := from.String()

			// Allow nil-able strings for optional fields.
			if fromStr == "" && toKind == reflect.Ptr {
				return nil, nil
			}

			switch to.Interface().(type) {
			case time.Time, *time.Time:
				val, err := time.Parse(dateLayout, fromStr)
				if toKind == reflect.Ptr {
					return &val, err
				}
				return val, err
			}

			// Remove leading zeros from int types so ParseInt won't assume it's an octal.
			if toKind >= reflect.Int && toKind <= reflect.Int64 {
				return strings.TrimLeft(from.String(), "0"), nil
			}
		}
		return from.Interface(), nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
		Result:           output,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(fields); err != nil {
		return err
	}

	return nil
}

func (t *Transaction) UnmarshalLines(lines [][]byte, output interface{}) error {
	t.Append(lines...)
	return t.Unmarshal(output)
}

func (t *Transaction) Marshal(input interface{}) error {
	output := map[string]string{}

	decodeHook := func(from reflect.Value, to reflect.Value, tags map[string]string) (interface{}, error) {
		if from.Kind() == reflect.Ptr && from.IsNil() {
			return "", nil
		}

		switch from.Interface().(type) {
		case time.Time, *time.Time:
			layout := time.RFC3339
			if val, ok := tags["time"]; ok {
				layout = val
			}
			if from.Kind() == reflect.Ptr {
				return from.Interface().(*time.Time).Format(layout), nil
			}
			return from.Interface().(time.Time).Format(layout), nil
		}

		return from.Interface(), nil
	}

	if err := structs.FillMap(input, output, "txn", decodeHook); err != nil {
		return err
	}

	if err := t.SetFields(output); err != nil {
		return err
	}

	return nil
}

func (t *Transaction) MarshalLines(input interface{}) ([][]byte, error) {
	if err := t.Marshal(input); err != nil {
		return [][]byte{}, err
	}
	return t.Records, nil
}
