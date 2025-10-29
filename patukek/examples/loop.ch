loop = patukek(times, function) {
	if times > 0 {
		function()
		loop(times-1, function)
	}
}

loop(5, patukek() { println("Hello World") })