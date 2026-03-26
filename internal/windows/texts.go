package windows

import "fmt"

func FoundUsersText(usersCount int) string {
	// Исключения для чисел от 11 до 14
	if usersCount%100 >= 11 && usersCount%100 <= 14 {
		return fmt.Sprintf("Найдено %d пользователей:", usersCount)
	}

	switch usersCount % 10 {
	case 1:
		return fmt.Sprintf("Найден %d пользователь:", usersCount)
	case 2, 3, 4:
		return fmt.Sprintf("Найдено %d пользователя:", usersCount)
	default:
		return fmt.Sprintf("Найдено %d пользователей:", usersCount)
	}
}
