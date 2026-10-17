package structs

import (
	"fmt"
	"reflect"
)

// [start named]

// Celsius and Fahrenheit share an underlying type (float64), but they
// are different types. The compiler will not mix them up for you.
type Celsius float64
type Fahrenheit float64

func toF(c Celsius) Fahrenheit { return Fahrenheit(c*9/5 + 32) }

func Example_namedTypes() {
	boiling := Celsius(100)
	fmt.Println(toF(boiling))
	// toF(Fahrenheit(212)) would not compile: wrong type.

	// Output:
	// 212
}

// [end named]

// [start literal]

type Address struct {
	City    string
	Country string
}

type User struct {
	ID      int
	Name    string
	Email   string
	Address Address // a struct field of struct type
	Admin   bool
}

func Example_literals() {
	u := User{
		ID:   1,
		Name: "Ana",
		Address: Address{
			City:    "Chennai",
			Country: "India",
		},
	} // fields you leave out get their zero values

	fmt.Println(u.Name, u.Address.City, u.Admin, u.Email == "")
	fmt.Printf("%+v\n", u.Address) // %+v prints field names

	var empty User // the zero value: every field zero
	fmt.Println(empty.ID, empty.Name == "")

	// Output:
	// Ana Chennai false true
	// {City:Chennai Country:India}
	// 0 true
}

// [end literal]

// [start copy]

func Example_copyAndCompare() {
	a := Address{City: "Chennai", Country: "India"}
	b := a // copies every field
	b.City = "Madurai"
	fmt.Println(a.City, b.City)

	// Structs whose fields are all comparable can use ==.
	fmt.Println(a == Address{City: "Chennai", Country: "India"})

	// Output:
	// Chennai Madurai
	// true
}

// [end copy]

// [start tags]

type Product struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price,omitempty" db:"price_inr"`
}

func Example_tags() {
	field, _ := reflect.TypeFor[Product]().FieldByName("Price")
	fmt.Println(field.Tag.Get("json"))
	fmt.Println(field.Tag.Get("db"))

	// Output:
	// price,omitempty
	// price_inr
}

// [end tags]

// [start anon]

func Example_anonymousStructs() {
	// Handy for one-off data, especially test tables.
	cases := []struct {
		in   string
		want int
	}{
		{"go", 2},
		{"gopher", 6},
	}
	for _, c := range cases {
		fmt.Println(len(c.in) == c.want)
	}

	point := struct{ X, Y int }{3, 4}
	fmt.Println(point.X + point.Y)

	// Output:
	// true
	// true
	// 7
}

// [end anon]
