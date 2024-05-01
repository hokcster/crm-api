package main

import "github.com/mrFokin/jrpc"

var (
	ErrorForbidden        = jrpc.NewError(403, "Недостаточно прав", nil)
	ErrorNotFound         = jrpc.NewError(404, "Объект не найден", nil)
	ErrorIsUsed           = jrpc.NewError(226, "Объект используется", nil)
	ErrorSplitsOverlapped = jrpc.NewError(700, "Обнаружено пересечение сплитов или тренировок", nil)
)
