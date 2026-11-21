package repository_test

import (
	"testing"

	"pos/internal/repository"
)

func TestRepositoryRequiresSaveAndKeepsCopies(t *testing.T) {
	repo := repository.NewMemoryRepository()
	course, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := course.AddStudent("s1", "Анна"); err != nil {
		t.Fatal(err)
	}
	beforeSave, _ := repo.Load()
	if len(beforeSave.Students()) != 0 {
		t.Fatal("изменение попало в репозиторий без Save")
	}
	if err := repo.Save(course); err != nil {
		t.Fatal(err)
	}
	if err := course.RemoveStudent("s1"); err != nil {
		t.Fatal(err)
	}
	afterSave, _ := repo.Load()
	if len(afterSave.Students()) != 1 {
		t.Fatal("репозиторий сохранил ссылку вместо копии")
	}
	if err := repo.Save(nil); err == nil {
		t.Fatal("ожидалась ошибка при сохранении nil")
	}
}
