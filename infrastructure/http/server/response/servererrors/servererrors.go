package servererrors

import "net/http"

/*
InternalServerError отправляет ответ с кодом 500.

IMPORTANT: Обратный прокси перехватывает ошибки и устанавливает заголовки.
Например,
"Cache-Control: no-store".
*/
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

/*
ServiceUnavailable отправляет ответ с кодом 503.

IMPORTANT: Обратный прокси перехватывает ошибки и устанавливает заголовки.
Например,
"Cache-Control: no-store",
"Retry-After: 5".
*/
func ServiceUnavailable(w http.ResponseWriter) {
	w.WriteHeader(http.StatusServiceUnavailable)
}
