package objects

type Class struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Week Week   `json:"week"`
}

type Week struct {
	Monday    []*Lesson `json:"monday"`
	Tuesday   []*Lesson `json:"tuesday"`
	Wednesday []*Lesson `json:"wednesday"`
	Thursday  []*Lesson `json:"thursday"`
	Friday    []*Lesson `json:"friday"`
}

type Teacher struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Lesson Lesson `json:"lesson"`
}

type Lesson struct {
	Name        string `json:"name"`
	TeacherName string `json:"teacher_name"`
}
