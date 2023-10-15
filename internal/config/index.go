package config

// Index allows fast lookup of endpoints by method and number of fragments
// It is read-only while the server is running, everything is built at startup
type Index map[uint64][]*Endpoint

// newIndex creates the tree from a config
func newIndex(api *API) Index {
	idx := make(Index)
	for _, endpoint := range api.Endpoints {
		h := hash(endpoint.Method, len(endpoint.Fragments))
		if idx[h] == nil {
			idx[h] = make([]*Endpoint, 0)
		}
		idx[h] = append(idx[h], endpoint)
	}
	return idx
}

// hash returns a unique id for a method and number of fragments
func hash(method string, lenFragments int) uint64 {
	switch method {
	case "GET":
		return uint64(1 + 10*lenFragments)
	case "POST":
		return uint64(2 + 10*lenFragments)
	case "PUT":
		return uint64(3 + 10*lenFragments)
	case "DELETE":
		return uint64(4 + 10*lenFragments)
	default:
		return uint64(10 * lenFragments)
	}
}

// Find the endpoint matching the method and fragments
func (t Index) Find(method string, fragments []string, validators Validators) *Endpoint {
	id := hash(method, len(fragments))
	endpoints := t[id]
	if endpoints == nil {
		return nil
	}
	for _, endpoint := range endpoints {
		if endpoint.Match(fragments, validators) {
			return endpoint
		}
	}
	return nil
}
