package internal

// Use to get colored text for printing to terminal.
// Returns "<color_code><str><clear>"
// Ex: Color_Test("Hello", "green") returns "\033[32mHello\033[0m"
func ColorText(str string, color string) string {
	// ANSI color escape sequences.
	var clear = "\033[0m"
	var green = "\033[32"
	var yellow = "\033[33"
	var reg_suffix = "m"

	var colStr string
	if color == "green" {
		colStr += green
	} else if color == "yellow" {
		colStr += yellow
	} else {
		//Invalid color provided.
		return str
	}

	colStr += reg_suffix + str + clear
	return colStr
}

// Adds commas to number string at hundreds place, thousands place, etc.
// Ex: "12345678" -> "12,345,678"
func AddCommas(str string) string {
	for i := len(str) - 3; i >= 0; i -= 3 {
		str = str[:i] + "," + str[i:]
	}
	return str
}
