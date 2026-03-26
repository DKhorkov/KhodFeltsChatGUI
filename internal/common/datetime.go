package common

import (
	"time"
	_ "time/tzdata" // нужно для подгрузки таймзон в докер контейнере
)

const (
	DateFormat     = "02.01.2006"
	DateTimeFormat = DateFormat + " 15:04:05"
)

// Timezone Получаем временную зону пользователя, чтобы отображать время сообщений по его местному времени.
var Timezone = time.Now().Location()
