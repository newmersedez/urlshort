package paniccheck

func panicInRegularFunc() {
	panic("something went wrong")
}

func panicInCondition(ok bool) {
	if !ok {
		panic("not ok")
	}
}
