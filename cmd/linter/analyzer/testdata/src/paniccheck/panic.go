package paniccheck

func panicInRegularFunc() {
	panic("something went wrong") // want "use of built-in panic is not allowed"
}

func panicInCondition(ok bool) {
	if !ok {
		panic("not ok") // want "use of built-in panic is not allowed"
	}
}
