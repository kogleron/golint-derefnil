package examples

type argStr struct {
	Vrbl int
}

func (a *argStr) getVrbl(def int) int {
	if a == nil {
		return def
	}

	return a.Vrbl
}

func fooExpectedPropDeref(a *argStr) int {
	return a.Vrbl // want "no nil check for the arg 'a' of 'fooExpectedPropDeref' before dereferencing"
}

func fooExpectedWholeDeref(a *argStr) argStr {
	return *a // want "no nil check for the arg 'a' of 'fooExpectedWholeDeref' before dereferencing"
}

func fooExpectedNotStruct(a *int) int {
	return *a // want "no nil check for the arg 'a' of 'fooExpectedNotStruct' before dereferencing"
}

func fooNotExpectedNotRef(a argStr) int {
	return a.Vrbl
}

func fooNotExpectedMethDeref(a *argStr) int {
	return a.getVrbl(-1)
}
