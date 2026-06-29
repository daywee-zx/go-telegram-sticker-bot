package api

// Checks if the name contains only latin, digits or underscores
func isValidName(name string) bool {
	for _, n := range []rune(name) {
		if !((n >= 48 && n <= 57) || (n >= 65 && n <= 90) || (n >= 97 && n <= 122) || n == 95) {
			return false
		}
	}
	return true
}
