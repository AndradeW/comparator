package comparator

// ValidationError indica que la entrada del usuario no es válida (ej: URL bloqueada).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// UpstreamError indica un fallo al interactuar con el servidor destino (petición o lectura del cuerpo).
type UpstreamError struct {
	Message string
}

func (e *UpstreamError) Error() string {
	return e.Message
}
