//go:build go1.27

package jsonv2

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/json/jsontext"
	"encoding/json/v2"
)

// [start basics]

type Item struct {
	Name  string   `json:"name"`
	Tags  []string `json:"tags"`
	Price int      `json:"price,omitzero"`
}

func Example_v2Defaults() {
	// Same struct tags, different defaults from v1:
	// nil slices encode as [] (not null), and keys come out as written.
	b, err := json.Marshal(Item{Name: "pen"})
	fmt.Println(string(b), err)

	// Unknown members are still accepted by default...
	var it Item
	fmt.Println(json.Unmarshal([]byte(`{"name":"pen","color":"blue"}`), &it))

	// ...but can be rejected with an option, without a Decoder.
	err = json.Unmarshal([]byte(`{"name":"pen","color":"blue"}`), &it, json.RejectUnknownMembers(true))
	fmt.Println(errors.Is(err, json.ErrUnknownName))

	// Output:
	// {"name":"pen","tags":[]} <nil>
	// <nil>
	// true
}

// [end basics]

// [start strict]

func Example_v2Strictness() {
	var v map[string]int

	// v1 silently kept the last duplicate key; v2 rejects duplicates.
	err := json.Unmarshal([]byte(`{"a":1,"a":2}`), &v)
	fmt.Println(err != nil)

	// Map output order is sorted with Deterministic, useful for golden tests.
	b, _ := json.Marshal(map[string]int{"b": 2, "a": 1, "c": 3}, json.Deterministic(true))
	fmt.Println(string(b))

	// Output:
	// true
	// {"a":1,"b":2,"c":3}
}

// [end strict]

// [start jsontext]

func Example_jsontext() {
	// jsontext works at the token level: fast, streaming, always valid JSON.
	dec := jsontext.NewDecoder(bytes.NewReader([]byte(`{"items":[1,2,3],"ok":true}`)))
	count := 0
	for {
		tok, err := dec.ReadToken()
		if err != nil {
			break // io.EOF at the end of input
		}
		if tok.Kind() == '0' { // '0' is the kind for numbers
			count++
		}
	}
	fmt.Println("numbers:", count)

	v := jsontext.Value(`{"b": 1, "a": [1,  2]}`)
	v.Compact() // also: Indent, Canonicalize
	fmt.Println(string(v))

	// Output:
	// numbers: 3
	// {"b":1,"a":[1,2]}
}

// [end jsontext]
