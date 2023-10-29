package responses

type Response struct {
	Data any
}

func NewResponse(data any) Response {
	return Response{Data: data}
}
