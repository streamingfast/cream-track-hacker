package main

type addressSet []string

func (s addressSet) contains(address string) bool {
	for _, candidate := range s {
		if candidate == address {
			return true
		}
	}

	return false
}
