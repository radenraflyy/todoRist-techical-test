package exception

type CustomException struct {
	Code    int
	Message string
}

func (e *CustomException) Error() string {
	return e.Message
}

type BadRequestException struct {
	Message string
}

func (e *BadRequestException) Error() string {
	return e.Message
}

type UnautorizedException struct {
	Message string
}

func (e *UnautorizedException) Error() string {
	return e.Message
}

type NotFoundException struct {
	Message string
}

func (e *NotFoundException) Error() string {
	return e.Message
}
