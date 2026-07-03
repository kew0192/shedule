package objects

import (
	"context"
)

type Repository interface {
	// Классы
	GetClass(ctx context.Context, id int) (*Class, error)
	GetAllClasses(ctx context.Context) ([]*Class, error)
	GetClassByName(ctx context.Context, name string) (*Class, error)
	UpdateClass(ctx context.Context, class *Class) error
	CreateClass(ctx context.Context, name string) (*Class, error)
	DeleteClass(ctx context.Context, id int) error

	// Учителя
	GetTeacher(ctx context.Context, id int) (*Teacher, error)
	GetTeacherByName(ctx context.Context, name string) (*Teacher, error)
	CreateTeacher(ctx context.Context, teacher *Teacher) error
	DeleteTeacher(ctx context.Context, id int) error
	GetTeacherByNameAndSubject(ctx context.Context, name string, subject string) (*Teacher, error)

	InitTables(ctx context.Context) error
}
