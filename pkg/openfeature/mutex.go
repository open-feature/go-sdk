package openfeature

// for the sake of testing
type mutex interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}
