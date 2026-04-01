package usecase

import (
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

type HabitUseCase struct {
	repo service.MemoryService
}

func NewHabitUseCase(repo service.MemoryService) *HabitUseCase {
	return &HabitUseCase{repo: repo}
}

func (uc *HabitUseCase) LogHabit(name string) error {
	return uc.repo.LogHabit(name)
}

func (uc *HabitUseCase) GetStreak(name string) (int, int, error) {
	streak, total, err := uc.repo.GetHabitStreak(name)
	if err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	return streak, total, nil
}

func (uc *HabitUseCase) ListToday() ([]string, error) {
	habits, err := uc.repo.ListHabitsToday()
	if err != nil {
		return nil, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	if habits == nil {
		habits = []string{}
	}
	return habits, nil
}
