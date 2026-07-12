package domains

type Reaction struct {
	ID    uint64 `json:"id"`
	Emoji string `json:"emoji"`
	// SortOrder — глобальный порядок отображения в UI (пикер + бэйджи).
	// Клиент сортирует по нему, чтобы порядок не рушился при асинхронных WS-событиях.
	SortOrder uint64 `json:"sortOrder"`
}

type MessageReactionSummary struct {
	Reaction Reaction `json:"reaction"`
	UserIDs  []uint64 `json:"userIds"`
}

type MessageReactionDTO struct {
	MessageID  uint64 `json:"-"`
	ReactionID uint64 `json:"reactionId"`
}
