package servers

type Http2Server struct {
}

func NewHttp2Server() *Http2Server {
	return &Http2Server{}
}

func (server *Http2Server) ListenAndServe(port int) {
	panic("http2 not implemented yet")
	//socket, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
}
