package main

// Transitions calculates the next state of the machine and the action to take
// given the current state of the machine and an event.
// It knows when a session should be locked, if the display should be off or on.
func Transition(currentState string, event string) (string, string) {
	switch {
	case currentState == "active" && event == "idle5m":
		return "locked", "lock"
	case currentState == "locked" && event == "idle10m":
		return "displayoff", "displayoff"
	case currentState == "displayoff" && event == "idle15m":
		return "displayoff", "suspend"
	case currentState == "displayoff" && event == "idle20m":
		return "displayoff", "hibernate"
	case currentState == "displayoff" && event == "userinput":
		return "locked", "displayon"
	}
	return "", ""
}
