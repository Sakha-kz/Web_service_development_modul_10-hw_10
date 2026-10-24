package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Book представляет книгу в библиотеке
type Book struct {
	ID        int
	Title     string
	Author    string
	Year      int
	Available int // доступно экземпляров
	Total     int // всего экземпляров
}

// Reader представляет зарегистрированного читателя
type Reader struct {
	ID   int
	Name string
}

// BorrowRecord фиксирует выдачу книги читателю
type BorrowRecord struct {
	ReaderID int
	BookID   int
}

// LibraryManager управляет книжным фондом, читателями и выдачами
type LibraryManager struct {
	books        []Book
	readers      map[int]Reader
	borrowed     []BorrowRecord
	borrowCounts map[int]int // BookID -> количество выдач за все время
	nextBookID   int
	nextReaderID int
}

func NewLibraryManager() *LibraryManager {
	lm := &LibraryManager{
		readers:      make(map[int]Reader),
		borrowCounts: make(map[int]int),
		nextBookID:   6,
		nextReaderID: 104,
	}

	// Начальные книги
	lm.books = []Book{
		{ID: 1, Title: "The Go Programming Language", Author: "Alan Donovan, Brian Kernighan", Year: 2015, Available: 3, Total: 5},
		{ID: 2, Title: "Clean Code", Author: "Robert C. Martin", Year: 2008, Available: 4, Total: 4},
		{ID: 3, Title: "Design Patterns (GoF)", Author: "Erich Gamma et al.", Year: 1994, Available: 2, Total: 3},
		{ID: 4, Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Year: 2017, Available: 1, Total: 2},
		{ID: 5, Title: "Introduction to Algorithms", Author: "Thomas H. Cormen", Year: 2009, Available: 5, Total: 5},
	}

	// Начальные читатели
	lm.readers[101] = Reader{ID: 101, Name: "Иван Петров"}
	lm.readers[102] = Reader{ID: 102, Name: "Анна Сидорова"}
	lm.readers[103] = Reader{ID: 103, Name: "Дмитрий Козлов"}

	// Фиксация предварительных выдач для книги 1
	lm.borrowed = append(lm.borrowed, BorrowRecord{ReaderID: 101, BookID: 1})
	lm.borrowed = append(lm.borrowed, BorrowRecord{ReaderID: 102, BookID: 1})
	lm.borrowCounts[1] = 5
	lm.borrowCounts[2] = 2
	lm.borrowCounts[3] = 3

	return lm
}

// 1. AddBook добавляет новую книгу
func (lm *LibraryManager) AddBook(title, author string, year, count int) {
	b := Book{
		ID:        lm.nextBookID,
		Title:     strings.TrimSpace(title),
		Author:    strings.TrimSpace(author),
		Year:      year,
		Available: count,
		Total:     count,
	}
	lm.books = append(lm.books, b)
	lm.nextBookID++
	fmt.Printf("Книга #%d ('%s') успешно добавлена в фонд!\n", b.ID, b.Title)
}

// 2. RegisterReader регистрирует читателя в карте
func (lm *LibraryManager) RegisterReader(name string) {
	r := Reader{
		ID:   lm.nextReaderID,
		Name: strings.TrimSpace(name),
	}
	lm.readers[r.ID] = r
	lm.nextReaderID++
	fmt.Printf("Читатель #%d (%s) успешно зарегистрирован!\n", r.ID, r.Name)
}

// FindBook находит книгу по ID и возвращает указатель на нее
func (lm *LibraryManager) FindBook(id int) *Book {
	for i := range lm.books {
		if lm.books[i].ID == id {
			return &lm.books[i]
		}
	}
	return nil
}

// 3. BorrowBook выдает книгу читателю по указателю
func (lm *LibraryManager) BorrowBook(book *Book, readerID int) error {
	if book == nil {
		return errors.New("книга не найдена")
	}
	if _, readerExists := lm.readers[readerID]; !readerExists {
		return fmt.Errorf("читатель с ID %d не зарегистрирован", readerID)
	}

	// Проверка: не взял ли уже читатель эту же книгу
	for _, rec := range lm.borrowed {
		if rec.ReaderID == readerID && rec.BookID == book.ID {
			return fmt.Errorf("читатель #%d уже взял экземпляр книги #%d", readerID, book.ID)
		}
	}

	if book.Available <= 0 {
		return fmt.Errorf("все экземпляры книги '%s' в данный момент на руках", book.Title)
	}

	book.Available--
	lm.borrowed = append(lm.borrowed, BorrowRecord{ReaderID: readerID, BookID: book.ID})
	lm.borrowCounts[book.ID]++
	return nil
}

// 4. ReturnBook принимает возврат книги от читателя по указателю
func (lm *LibraryManager) ReturnBook(book *Book, readerID int) error {
	if book == nil {
		return errors.New("книга не найдена")
	}

	foundIndex := -1
	for i, rec := range lm.borrowed {
		if rec.ReaderID == readerID && rec.BookID == book.ID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("читатель #%d не брал книгу '%s' (возврат невозможен)", readerID, book.Title)
	}

	// Удаление из среза выданных
	lm.borrowed = append(lm.borrowed[:foundIndex], lm.borrowed[foundIndex+1:]...)
	book.Available++
	return nil
}

// 5. SearchBooks ищет книги по названию или автору
func (lm *LibraryManager) SearchBooks(query string) []Book {
	var result []Book
	q := strings.ToLower(strings.TrimSpace(query))
	for _, b := range lm.books {
		if strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.Author), q) {
			result = append(result, b)
		}
	}
	return result
}

// 6. GetReaderBooks возвращает список книг, находящихся у читателя
func (lm *LibraryManager) GetReaderBooks(readerID int) []Book {
	var list []Book
	for _, rec := range lm.borrowed {
		if rec.ReaderID == readerID {
			if b := lm.FindBook(rec.BookID); b != nil {
				list = append(list, *b)
			}
		}
	}
	return list
}

// 7. GetStatistics формирует общую статистику библиотеки
func (lm *LibraryManager) GetStatistics() (titles, totalCopies, available, borrowed int, topBookTitle string, topBorrows int) {
	titles = len(lm.books)
	for _, b := range lm.books {
		totalCopies += b.Total
		available += b.Available
	}
	borrowed = len(lm.borrowed)

	maxB := 0
	topID := 0
	for id, cnt := range lm.borrowCounts {
		if cnt > maxB {
			maxB = cnt
			topID = id
		}
	}

	if b := lm.FindBook(topID); b != nil {
		topBookTitle = b.Title
		topBorrows = maxB
	}
	return
}

func main() {
	lm := NewLibraryManager()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============ LIBRARY MANAGER ============")
		fmt.Println("1. Добавить книгу")
		fmt.Println("2. Зарегистрировать читателя")
		fmt.Println("3. Выдать книгу")
		fmt.Println("4. Вернуть книгу")
		fmt.Println("5. Найти книгу (поиск)")
		fmt.Println("6. Книги читателя")
		fmt.Println("7. Статистика библиотеки")
		fmt.Println("0. Выход")
		fmt.Print("Выберите действие: ")

		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Print("Название книги: ")
			scanner.Scan()
			title := scanner.Text()

			fmt.Print("Автор(ы): ")
			scanner.Scan()
			author := scanner.Text()

			fmt.Print("Год издания: ")
			scanner.Scan()
			year, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

			fmt.Print("Количество экземпляров: ")
			scanner.Scan()
			count, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || count <= 0 {
				fmt.Println("Ошибка: количество должно быть > 0!")
				continue
			}

			lm.AddBook(title, author, year, count)

		case "2":
			fmt.Print("ФИО читателя: ")
			scanner.Scan()
			name := scanner.Text()
			lm.RegisterReader(name)

		case "3":
			fmt.Print("ID читателя: ")
			scanner.Scan()
			rID, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil {
				fmt.Println("Некорректный ID читателя.")
				continue
			}

			fmt.Print("ID книги: ")
			scanner.Scan()
			bID, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil {
				fmt.Println("Некорректный ID книги.")
				continue
			}

			bookPtr := lm.FindBook(bID)
			if err := lm.BorrowBook(bookPtr, rID); err != nil {
				fmt.Printf("Ошибка выдачи: %v\n", err)
			} else {
				reader := lm.readers[rID]
				fmt.Println("\nКнига успешно выдана!")
				fmt.Printf("Название:             %s\n", bookPtr.Title)
				fmt.Printf("Читатель:             %s\n", reader.Name)
				fmt.Printf("Осталось экземпляров: %d\n", bookPtr.Available)
			}

		case "4":
			fmt.Print("ID читателя: ")
			scanner.Scan()
			rID, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

			fmt.Print("ID книги: ")
			scanner.Scan()
			bID, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

			bookPtr := lm.FindBook(bID)
			if err := lm.ReturnBook(bookPtr, rID); err != nil {
				fmt.Printf("Ошибка возврата: %v\n", err)
			} else {
				fmt.Printf("Книга '%s' успешно возвращена! Доступно: %d шт.\n", bookPtr.Title, bookPtr.Available)
			}

		case "5":
			fmt.Print("Поисковый запрос (название или автор): ")
			scanner.Scan()
			q := scanner.Text()
			matched := lm.SearchBooks(q)
			fmt.Printf("\nНайдено книг: %d\n", len(matched))
			for _, b := range matched {
				fmt.Printf("- #%d: '%s' (%s, %d г.) | Доступно: %d из %d\n",
					b.ID, b.Title, b.Author, b.Year, b.Available, b.Total)
			}

		case "6":
			fmt.Print("ID читателя: ")
			scanner.Scan()
			rID, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			r, exists := lm.readers[rID]
			if !exists {
				fmt.Println("Читатель не найден.")
				continue
			}
			books := lm.GetReaderBooks(rID)
			fmt.Printf("\nКниги на руках у читателя %s (ID %d): %d шт.\n", r.Name, rID, len(books))
			for _, b := range books {
				fmt.Printf("- #%d: '%s' (%s)\n", b.ID, b.Title, b.Author)
			}

		case "7":
			titles, copies, avail, borrowed, topTitle, topCnt := lm.GetStatistics()
			fmt.Println("\n--- СТАТИСТИКА БИБЛИОТЕКИ ---")
			fmt.Printf("Всего наименований:       %d\n", titles)
			fmt.Printf("Всего экземпляров:        %d\n", copies)
			fmt.Printf("Доступно:                 %d\n", avail)
			fmt.Printf("Выдано на руки:           %d\n", borrowed)
			fmt.Println("Самая популярная книга:")
			fmt.Printf("'%s' — количество выдач: %d\n", topTitle, topCnt)

		case "0":
			fmt.Println("Завершение работы Library Manager.")
			return

		default:
			fmt.Println("Неверный пункт меню. Введите число от 0 до 7.")
		}
	}
}
