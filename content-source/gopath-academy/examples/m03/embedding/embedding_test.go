package embedding

import (
	"fmt"
	"sync"
)

// [start promote]

type Person struct {
	Name  string
	Email string
}

// Employee embeds Person: it writes the type name with no field name.
type Employee struct {
	Person // embedded field; its name is "Person"
	Team   string
}

func Example_promotedFields() {
	e := Employee{
		Person: Person{Name: "Ana", Email: "ana@example.com"},
		Team:   "platform",
	}
	fmt.Println(e.Name)        // promoted from Person
	fmt.Println(e.Person.Name) // the explicit path still works
	fmt.Println(e.Team)

	// Output:
	// Ana
	// Ana
	// platform
}

// [end promote]

// [start mutex]

// Embedding a type also promotes its methods. Counter gets Lock and
// Unlock directly from sync.Mutex (methods are covered in Module 4).
type Counter struct {
	sync.Mutex
	n int
}

func Example_promotedMethods() {
	var c Counter // zero value: unlocked mutex, n == 0
	c.Lock()
	c.n++
	c.Unlock()
	fmt.Println(c.n)

	// Output:
	// 1
}

// [end mutex]

// [start conflict]

type Timestamps struct {
	Name string // also called Name
	At   string
}

type Record struct {
	Person
	Timestamps
	Name string // depth 0: wins over both embedded Name fields
}

type Ambiguous struct {
	Person
	Timestamps
}

func Example_nameConflicts() {
	r := Record{Name: "outer", Person: Person{Name: "person"}}
	fmt.Println(r.Name, r.Person.Name)

	var a Ambiguous
	// a.Name would not compile: ambiguous selector a.Name
	a.Person.Name = "explicit"
	fmt.Println(a.Person.Name)

	// Output:
	// outer person
	// explicit
}

// [end conflict]

// [start notinheritance]

func describe(p Person) string { return p.Name + " <" + p.Email + ">" }

func Example_notInheritance() {
	e := Employee{Person: Person{Name: "Bo", Email: "bo@example.com"}}

	// An Employee is not a Person: describe(e) does not compile.
	// You pass the embedded value explicitly.
	fmt.Println(describe(e.Person))

	// Output:
	// Bo <bo@example.com>
}

// [end notinheritance]
