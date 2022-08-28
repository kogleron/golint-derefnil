package example

func (e *Example) foo() string { // want "no nil check for the receiver 'e' of 'foo' before accessing 'nameMe'"
	return e.nameMe
}

func (e *Example) koo() string {
	if e == nil {
		return ""
	}

	return e.nameMe
}

func (e *Example) boo() {
	e.foo()
}

func (e *Example) nope() {
}

type Example struct {
	nameMe      string
	anotherName int
}
