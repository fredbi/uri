// nolint: forbidigo
package uri_test

import (
	"fmt"

	"github.com/fredbi/uri"
)

func ExampleParse() {
	u, err := uri.Parse("https://example.com:8080/path")
	if err != nil {
		fmt.Println("Invalid URI:", err)
	} else {
		fmt.Println(u.String())
	}

	// Output: https://example.com:8080/path
}

func ExampleURI_Scheme() {
	u, err := uri.Parse("https://example.com:8080/path")
	if err != nil {
		fmt.Println("Invalid URI:", err)
	} else {
		fmt.Println(u.Scheme())
	}

	// Output: https
}

func ExampleURI_Authority() {
	u, err := uri.Parse("ftp://example.com/path")
	if err != nil {
		fmt.Println("Invalid URI:", err)
	} else {
		fmt.Println(u.Authority().Path())
	}
	// Output:
	// /path
}

func ExampleAuthority_Path() {
	u, err := uri.ParseReference("//example.com/path")
	if err != nil {
		fmt.Println("Invalid URI reference:", err)
	} else {
		fmt.Println(u.Authority().Path())
	}

	// Output:
	// /path
}

func ExampleParseReference() {
	u, err := uri.ParseReference("//example.com/path?a=1#fragment")
	if err != nil {
		fmt.Println("Invalid URI reference:", err)
	} else {
		fmt.Println(u.Fragment())

		params := u.Query()
		fmt.Println(params.Get("a"))
	}
	// Output:
	// fragment
	// 1
}

func ExampleIsURI() {
	isValid := uri.IsURI("urn://example.com?query=x#fragment/path") // true
	fmt.Println(isValid)

	isValid = uri.IsURI("//example.com?query=x#fragment/path") // false
	fmt.Println(isValid)

	// Output:
	// true
	// false
}

func ExampleIsURIReference() {
	isValid := uri.IsURIReference("//example.com?query=x#fragment/path") // true
	fmt.Println(isValid)

	// Output:
	// true
}
