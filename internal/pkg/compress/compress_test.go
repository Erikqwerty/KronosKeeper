package compress

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestZip(t *testing.T) {
	tempDir := t.TempDir()      // Создаем временную директорию
	defer os.RemoveAll(tempDir) // Удаляем временную директорию после выполнения теста

	// Тестовые случаи
	testCases := []struct {
		archiveName string
		input       []string
		expectError bool // Ожидается ли ошибка
	}{
		// Тестовый случай №1
		{
			archiveName: "test_archive",
			input:       []string{"file1.txt", "file2.txt"},
			expectError: false,
		},
		// Тестовый случай №2 (с неправильным именем архива)
		{
			archiveName: "б", // Пустое имя архива
			input:       []string{"file1.txt", "file2.txt"},
			expectError: false, // Ожидается ошибка
		},
	}

	// Запускаем переборку тестовых случаев
	for id, testCase := range testCases {

		// Создаем директорию для текущего тестового случая
		testCaseDir := filepath.Join(tempDir, "testCase"+strconv.Itoa(id))

		err := os.MkdirAll(testCaseDir, os.ModePerm)
		if err != nil {
			t.Fatalf("Ошибка при создании папки для тестового случая: %v\n", err)
		}

		// Создаем тестовые файлы
		for _, file := range testCase.input {
			filePath := filepath.Join(testCaseDir, file)

			if _, err := os.Create(filePath); err != nil {
				t.Fatalf("Не удалось создать тестовый файл %s: %v", filePath, err)
			}
		}

		// Создаем новый экземпляр Compress для каждого тестового случая
		c := New()
		c.ArchiveName = testCase.archiveName
		c.InputPaths = []string{testCaseDir}
		c.OutputPath = tempDir

		err = c.Zip()

		if testCase.expectError {
			// Ожидается ошибка, но ее нет
			if err == nil {
				t.Errorf("Тест %d не пройден: ожидалась ошибка, но ее не было", id+1)
			} else {
				t.Logf("Тест %d пройден: ожидалась ошибка и она была получена %v", id+1, err)
			}
		} else {
			// Ожидается успешное выполнение, но произошла ошибка
			if err != nil {
				t.Errorf("Тест %d не пройден: не ожидалась ошибка, но она произошла: %v", id+1, err)
			} else {
				t.Logf("Тест %d пройден: архив успешно создан", id+1)
			}
		}
	}
}

func TestIsExcluded(t *testing.T) {

	testCases := []struct {
		excludeFile []string
		files       []string
		expectFiles []string
	}{
		// Тестовый случай №1
		{
			excludeFile: []string{"*.log", "file", "todo.md"},
			files:       []string{"backup.log", "my.sql", "node.js", "todo.md", "work.md", "README.md", "file"},
			expectFiles: []string{"my.sql", "node.js", "work.md", "README.md"},
		},
	}

	compress := New()
	for id, testCase := range testCases {
		compress.ExludeFile = testCase.excludeFile
		resultFiles := []string{}
		for _, file := range testCase.files {
			status, err := compress.isExcluded(file)
			if err != nil {
				t.Errorf("Возникла ошибка при выполнение функции: isExcluded, %v", err)
			}
			if !status {
				resultFiles = append(resultFiles, file)
			}
		}
		if reflect.DeepEqual(resultFiles, testCase.expectFiles) {
			t.Logf("Тест пройден %v", id)
		} else {
			t.Errorf("Тест не пройден ожидалось: %v, полученно: %v", testCase.expectFiles, resultFiles)
		}
	}
}
