package http3

type Http3Server struct {
}

func NewHttp3Server() *Http3Server {
	return &Http3Server{}
}

func (server *Http3Server) ListenAndServe(port int) {
	panic("http3 not implemented yet")
	//socket, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
}
