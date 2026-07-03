package objects

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) AddLessons(ctx context.Context, teacher Teacher, class Class, hours int) (*Class, error) {
	remaining := hours

	days := []struct {
		name string
		day  []*Lesson
	}{
		{"monday", class.Week.Monday},
		{"tuesday", class.Week.Tuesday},
		{"wednesday", class.Week.Wednesday},
		{"thursday", class.Week.Thursday},
		{"friday", class.Week.Friday},
	}

	allClasses, err := s.Repo.GetAllClasses(ctx)
	if err != nil {
		return nil, errors.New("failed to get classes for teacher availability check")
	}

	// Проходим по дням несколько раз, чтобы найти свободные слоты
	for remaining > 0 {
		addedInRound := 0

		for _, d := range days {
			if remaining == 0 {
				break
			}

			// Проверяем, есть ли уже такой предмет в этот день
			if s.lessonExistsInDay(d.day, teacher.Lesson.Name) {
				continue
			}

			// Проверяем технические предметы
			if s.isTechnicalSubject(teacher.Lesson.Name) {
				techCount := s.countTechnicalInDay(d.day)
				if techCount >= 3 {
					continue
				}
			}

			// Ищем первый свободный слот
			for i := 0; i < len(d.day); i++ {
				if d.day[i] == nil {
					// Проверяем доступность учителя
					if s.isTeacherAvailable(allClasses, teacher.Name, d.name, i, class.ID) {
						d.day[i] = &Lesson{
							Name:        teacher.Lesson.Name,
							TeacherName: teacher.Name,
						}
						remaining--
						addedInRound++
						break
					}
				}
			}
		}

		// Если за проход не добавили ни одного урока - выходим
		if addedInRound == 0 {
			break
		}
	}

	if remaining > 0 {
		return nil, fmt.Errorf("not enough free slots or days, remaining: %d", remaining)
	}

	return &class, nil
}

// Проверяет, есть ли уже такой предмет в день
func (s *Service) lessonExistsInDay(day []*Lesson, lessonName string) bool {
	for _, slot := range day {
		if slot != nil && slot.Name == lessonName {
			return true
		}
	}
	return false
}

// Считает количество технических предметов в день
func (s *Service) countTechnicalInDay(day []*Lesson) int {
	count := 0
	for _, slot := range day {
		if slot != nil && s.isTechnicalSubject(slot.Name) {
			count++
		}
	}
	return count
}

// Проверяет доступность учителя в конкретный день и час
func (s *Service) isTeacherAvailable(allClasses []*Class, teacherName string, day string, slotIndex int, excludeClassID int) bool {
	for _, class := range allClasses {
		// Пропускаем текущий класс
		if class.ID == excludeClassID {
			continue
		}

		var daySlots []*Lesson
		switch day {
		case "monday":
			daySlots = class.Week.Monday
		case "tuesday":
			daySlots = class.Week.Tuesday
		case "wednesday":
			daySlots = class.Week.Wednesday
		case "thursday":
			daySlots = class.Week.Thursday
		case "friday":
			daySlots = class.Week.Friday
		default:
			return false
		}

		// Проверяем, занят ли учитель в этот час
		if slotIndex < len(daySlots) && daySlots[slotIndex] != nil &&
			daySlots[slotIndex].TeacherName == teacherName {
			return false
		}
	}
	return true
}

func (s *Service) isTechnicalSubject(name string) bool {
	techSubjects := map[string]bool{
		"Математика":  true,
		"Геометрия":   true,
		"Алгебра":     true,
		"Физика":      true,
		"Информатика": true,
		"Химия":       true,
		"Биология":    true,
	}
	return techSubjects[name]
}

func NewClass(ctx context.Context, name string) (*Class, error) {
	week := Week{
		Monday:    make([]*Lesson, 8),
		Tuesday:   make([]*Lesson, 8),
		Wednesday: make([]*Lesson, 8),
		Thursday:  make([]*Lesson, 8),
		Friday:    make([]*Lesson, 8),
	}

	// Извлекаем номер класса
	num := 0
	if len(name) > 0 {
		if name[0] >= '0' && name[0] <= '9' {
			num = int(name[0] - '0')
		}
		if len(name) > 1 && name[1] >= '0' && name[1] <= '9' {
			num = num*10 + int(name[1]-'0')
		}
	}

	if num >= 6 {
		week.Monday[0] = &Lesson{
			Name:        "Уроки о важном",
			TeacherName: "Классный руководитель",
		}

		week.Thursday[7] = &Lesson{
			Name:        "Мои горизонты",
			TeacherName: "Классный руководитель",
		}
	}

	return &Class{
		Name: name,
		Week: week,
	}, nil
}
