package skills

func FindTriggeredSkills(allSkills []Skill, completedTool string) []Skill {
	var triggered []Skill
	for _, s := range allSkills {
		if s.AfterTool != "" && s.AfterTool == completedTool && s.IsEnabled() {
			triggered = append(triggered, s)
		}
	}
	return triggered
}
