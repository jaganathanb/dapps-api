package common

func RecoverFromPanic(cb func(err any)) {
	if r := recover(); r != nil {
		if cb != nil {
			cb(r)
		}
	}
}
