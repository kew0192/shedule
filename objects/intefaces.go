package objects

import "context"

type LessonsAdder interface {
	AddLessons(ctx context.Context, teacher Teacher, Class Class, hours int) (*Class, error)
}

type LessonDeleter interface {
	DeleteLessons(ctx context.Context, teacher Teacher, Class Class) (*Class, error)
}

type ClassAdder interface {
	AddClass(ctx context.Context, name string) *Class
}
