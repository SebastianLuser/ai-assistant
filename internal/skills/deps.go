package skills

type DependencyChecker struct {
	available map[string]bool
}

func NewDependencyChecker(available map[string]bool) *DependencyChecker {
	return &DependencyChecker{available: available}
}

func (dc *DependencyChecker) AllAvailable(deps []string) bool {
	for _, d := range deps {
		if !dc.available[d] {
			return false
		}
	}
	return true
}

func (dc *DependencyChecker) Unavailable(deps []string) []string {
	var missing []string
	for _, d := range deps {
		if !dc.available[d] {
			missing = append(missing, d)
		}
	}
	return missing
}
