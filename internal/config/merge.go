package config

// merge lays tier over base: maps merge by key, anything else replaces. A null
// takes the value from defaults (the built-in tree at the same level), or
// removes the key. The result shares nothing with the inputs.
func merge(base, tier, defaults any) any {
	t, ok := tier.(map[string]any)
	if !ok {
		return clone(tier)
	}
	// base not a map → b is nil → tier merges onto empty, nulls still resolve.
	b, _ := base.(map[string]any)
	d, _ := defaults.(map[string]any)

	out := make(map[string]any, len(b)+len(t))
	for k, v := range b {
		if _, over := t[k]; !over {
			out[k] = clone(v)
		}
	}
	for k, v := range t {
		if v != nil {
			out[k] = merge(b[k], v, d[k])
		} else if dv, ok := d[k]; ok {
			out[k] = clone(dv)
		}
	}
	return out
}

// clone deep-copies a tree.
func clone(v any) any {
	switch v := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(v))
		for k, x := range v {
			m[k] = clone(x)
		}
		return m
	case []any:
		s := make([]any, len(v))
		for i, x := range v {
			s[i] = clone(x)
		}
		return s
	default:
		return v
	}
}
