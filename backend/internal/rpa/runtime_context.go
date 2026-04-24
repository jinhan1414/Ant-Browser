package rpa

type RuntimeContext struct {
	profileID string
	values    map[string]any
}

func NewRuntimeContext(profileID string) *RuntimeContext {
	return &RuntimeContext{
		profileID: profileID,
		values: map[string]any{
			"profileId": profileID,
		},
	}
}

func (c *RuntimeContext) Clone() *RuntimeContext {
	if c == nil {
		return NewRuntimeContext("")
	}
	next := NewRuntimeContext(c.profileID)
	for key, value := range c.values {
		next.values[key] = value
	}
	return next
}

func (c *RuntimeContext) Set(name string, value any) {
	if c == nil || name == "" {
		return
	}
	c.values[name] = value
}

func (c *RuntimeContext) Get(name string) any {
	if c == nil || name == "" {
		return nil
	}
	return c.values[name]
}
