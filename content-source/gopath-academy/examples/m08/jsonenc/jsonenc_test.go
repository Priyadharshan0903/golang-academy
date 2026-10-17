package jsonenc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// [start marshal]

type Address struct {
	City string `json:"city"`
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"` // left out when empty
	Address   *Address  `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"-"` // never encoded
	internal  int       // unexported: invisible to encoding/json
}

func Example_marshal() {
	u := User{
		ID: 1, Name: "Ana",
		CreatedAt: time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC),
		Password:  "hunter2", internal: 42,
	}
	b, err := json.Marshal(u)
	fmt.Println(string(b), err)

	pretty, _ := json.MarshalIndent(Address{City: "Chennai"}, "", "  ")
	fmt.Println(string(pretty))

	// Output:
	// {"id":1,"name":"Ana","created_at":"2026-10-16T09:30:00Z"} <nil>
	// {
	//   "city": "Chennai"
	// }
}

// [end marshal]

// [start unmarshal]

func Example_unmarshal() {
	data := []byte(`{"id": 7, "name": "Bo", "address": {"city": "Madurai"}, "extra": true}`)

	var u User
	if err := json.Unmarshal(data, &u); err != nil { // pass a POINTER
		fmt.Println(err)
		return
	}
	fmt.Println(u.ID, u.Name, u.Address.City) // "extra" is silently ignored

	var bad User
	err := json.Unmarshal([]byte(`{"id": "seven"}`), &bad)
	var typeErr *json.UnmarshalTypeError
	fmt.Println(errors.As(err, &typeErr), typeErr.Field)

	// Output:
	// 7 Bo Madurai
	// true id
}

// [end unmarshal]

// [start strict]

func Example_strictDecoding() {
	dec := json.NewDecoder(strings.NewReader(`{"id": 1, "nmae": "typo"}`))
	dec.DisallowUnknownFields() // reject fields the struct does not have

	var u User
	err := dec.Decode(&u)
	fmt.Println(err != nil)

	// Output:
	// true
}

// [end strict]

// [start stream]

func Example_streaming() {
	// Newline-delimited JSON (one object per line), common for logs and APIs.
	input := `{"id":1,"name":"a"}
{"id":2,"name":"b"}
{"id":3,"name":"c"}`
	dec := json.NewDecoder(strings.NewReader(input))
	sum := 0
	for dec.More() {
		var u User
		if err := dec.Decode(&u); err != nil {
			fmt.Println(err)
			return
		}
		sum += u.ID
	}
	fmt.Println(sum)

	enc := json.NewEncoder(os.Stdout) // writes one value per line
	enc.Encode(map[string]int{"total": sum})

	// Output:
	// 6
	// {"total":6}
}

// [end stream]

// [start custom]

type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusSuspended
)

var statusNames = map[Status]string{StatusActive: "active", StatusSuspended: "suspended"}

// MarshalText makes Status encode as a string, in JSON and anywhere
// else that understands encoding.TextMarshaler (also as map keys).
func (s Status) MarshalText() ([]byte, error) {
	name, ok := statusNames[s]
	if !ok {
		return nil, fmt.Errorf("unknown status %d", int(s))
	}
	return []byte(name), nil
}

func (s *Status) UnmarshalText(b []byte) error {
	for k, v := range statusNames {
		if v == string(b) {
			*s = k
			return nil
		}
	}
	return fmt.Errorf("unknown status %q", b)
}

type Account struct {
	Status Status `json:"status"`
}

func Example_customEncoding() {
	b, _ := json.Marshal(Account{Status: StatusSuspended})
	fmt.Println(string(b))

	var a Account
	err := json.Unmarshal([]byte(`{"status":"active"}`), &a)
	fmt.Println(a.Status == StatusActive, err)

	// Output:
	// {"status":"suspended"}
	// true <nil>
}

// [end custom]

// [start raw]

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"` // decoded later, once we know the type
}

type Login struct {
	User string `json:"user"`
}

type Purchase struct {
	Amount int `json:"amount"`
}

func Example_rawMessage() {
	stream := `[{"type":"login","payload":{"user":"ana"}},{"type":"purchase","payload":{"amount":499}}]`
	var events []Event
	json.Unmarshal([]byte(stream), &events)

	for _, e := range events {
		switch e.Type {
		case "login":
			var l Login
			json.Unmarshal(e.Payload, &l)
			fmt.Println("login by", l.User)
		case "purchase":
			var p Purchase
			json.Unmarshal(e.Payload, &p)
			fmt.Println("purchase of", p.Amount)
		}
	}

	// Output:
	// login by ana
	// purchase of 499
}

// [end raw]

// [start numbers]

func Example_numbersInAny() {
	var v map[string]any
	json.Unmarshal([]byte(`{"count": 3, "id": 12345678901234567890}`), &v)
	fmt.Printf("%T\n", v["count"]) // every JSON number becomes float64

	dec := json.NewDecoder(strings.NewReader(`{"id": 12345678901234567890}`))
	dec.UseNumber() // keep the exact text instead
	var exact map[string]any
	dec.Decode(&exact)
	fmt.Println(exact["id"])

	// Output:
	// float64
	// 12345678901234567890
}

// [end numbers]

// [start omitzero]

type Window struct {
	Start time.Time `json:"start,omitzero"` // Go 1.24+: omitted when Start.IsZero()
	End   time.Time `json:"end,omitempty"`  // omitempty never omits a struct
}

func Example_omitZero() {
	b, _ := json.Marshal(Window{})
	fmt.Println(string(b))

	// Output:
	// {"end":"0001-01-01T00:00:00Z"}
}

// [end omitzero]
