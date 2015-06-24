package expressionnotused

func F() {
	(4 & (1 << 1)) != 0
	noop(1 + 1)
}

func noop(a int) {
}
