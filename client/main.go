package client

// Get retrieves a value of an SKVS key
// from a dockerized instance
func Get(key string) (string, error) {
	c, err := NewFromDocker()
	if err != nil {
		return "", err
	}
	return c.Get(key)
}

// Set sets a value of a given SKVS key
// in a dockerized instance
func Set(key string, value string) error {
	c, err := NewFromDocker()
	if err != nil {
		return err
	}
	return c.Set(key, value)
}

// Delete removes an SKVS entry and all its children
// in a dockerized instance
func Delete(key string) error {
	c, err := NewFromDocker()
	if err != nil {
		return err
	}
	return c.Delete(key)
}
