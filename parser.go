package fixedwidth

// Unmarshal lines of byte strings into the given output struct.
func Unmarshal(lines [][]byte, output interface{}) error {
	t, err := NewTransaction(output)
	if err != nil {
		return err
	}
	t.Append(lines...)
	return t.Unmarshal(output)
}

// Marshal the given input struct to lines of byte strings.
func Marshal(input interface{}) ([][]byte, error) {
	t, err := NewTransaction(input)
	if err != nil {
		return [][]byte{}, err
	}
	if err := t.Marshal(input); err != nil {
		return [][]byte{}, err
	}
	return t.Records, nil
}
