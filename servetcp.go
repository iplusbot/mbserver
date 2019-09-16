package mbserver

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}

		go func(conn net.Conn) {
			defer conn.Close()

			header := make([]byte, 8)

			for {
				n, err := conn.Read(header)
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v\n", err)
					}
					return
				}
				if n != 8 {
					log.Printf("malformed header, expect length:8, got length:%d, data:%v\n", n, header)
					return
				}
				len := int(binary.BigEndian.Uint16(header[4:6]))
				packet := make([]byte, 8+len-2)
				copy(packet, header)
				n, err = conn.Read(packet[8:])
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v\n", err)
					}
					return
				}
				if n != len-2 {
					log.Printf("malformed data, expect length:%d, got length:%d, data:%v\n", n, len-2, packet)
					return
				}
				// Set the length of the packet to the number of read bytes.
				frame, err := NewTCPFrame(packet)
				if err != nil {
					log.Printf("bad packet error %v\n", err)
					return
				}

				request := &Request{conn, frame}

				s.requestChan <- request
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}
