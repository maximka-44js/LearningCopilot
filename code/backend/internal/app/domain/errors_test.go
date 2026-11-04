package domain

import "testing"

func TestNewValidationErrorUsesLocalizedMessages(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		message string
		want    string
	}{
		{
			code:    "project.deadline_invalid",
			message: "Project deadline cannot be before created_at",
			want:    "Дедлайн проекта должен быть корректной датой и не может быть раньше даты создания проекта.",
		},
		{
			code:    "task.difficulty.invalid",
			message: "Task difficulty must be between 1 and 10",
			want:    "Сложность задачи должна быть числом от 1 до 10.",
		},
		{
			code:    "folder.name_invalid",
			message: "Folder name contains unsupported characters",
			want:    "Укажите корректное название папки.",
		},
	}

	for _, test := range tests {
		t.Run(string(test.code), func(t *testing.T) {
			got := NewValidationError(test.code, test.message).Message
			if got != test.want {
				t.Fatalf("message = %q, want %q", got, test.want)
			}
		})
	}
}
