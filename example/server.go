
// // Interface to listen on a TFO enabled TCP socket
// package main

// import (
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net"
// 	"syscall"
// 	"os"
// 	"io"
	
// 	// "net/http"
// )

// type TFOServer struct {
// 	ServerAddr [4]byte
// 	ServerPort int
// 	fd         int
// }

// const TCP_FASTOPEN int = 23
// const LISTEN_BACKLOG int = 23

// // Create a tcp socket, setting the TCP_FASTOPEN socket option.
// func (s *TFOServer) Bind() (err error) {

// 	s.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
// 	if err != nil {
// 		if err == syscall.ENOPROTOOPT {
// 			err = errors.New("TCP Fast Open server support is unavailable (unsupported kernel).")
// 		}
// 		return
// 	}

// 	err = syscall.SetsockoptInt(s.fd, syscall.SOL_TCP, TCP_FASTOPEN, 1)
// 	if err != nil {
// 		err = errors.New(fmt.Sprintf("Failed to set necessary TCP_FASTOPEN socket option: %s", err))
// 		return
// 	}

// 	sa := &syscall.SockaddrInet4{Addr: s.ServerAddr, Port: s.ServerPort}

// 	err = syscall.Bind(s.fd, sa)
// 	if err != nil {
// 		err = errors.New(fmt.Sprintf("Failed to bind to Addr: %v, Port: %d, Reason: %s", s.ServerAddr, s.ServerPort, err))
// 		return
// 	}

// 	log.Printf("Server: Bound to addr: %v, port: %d\n", s.ServerAddr, s.ServerPort)

// 	err = syscall.Listen(s.fd, LISTEN_BACKLOG)
// 	if err != nil {
// 		err = errors.New(fmt.Sprintf("Failed to listen: %s", err))
// 		return
// 	}

// 	return

// }

// // Block, waiting for connections, handling each connection in its own go
// // routine
// func (s *TFOServer) Accept() {

// 	log.Println("Server: Waiting for connections")

// 	defer syscall.Close(s.fd)

// 	for {

// 		fd, sockaddr, err := syscall.Accept(s.fd)
// 		if err != nil {
// 			log.Fatalln("Failed to accept(): ", err)
// 		}

// 		cxn := TFOServerConn{fd: fd, sockaddr: sockaddr.(*syscall.SockaddrInet4)}

// 		go cxn.Handle()

// 	}

// }

// // A client/server connection accepted by TFOServer
// type TFOServerConn struct {
// 	sockaddr *syscall.SockaddrInet4
// 	fd       int
// }

// func (cxn *TFOServerConn) Write(p []byte) (n int, err error) {
//     return syscall.Write(cxn.fd, p)
// }

// func (cxn *TFOServerConn) Handle() {
//     defer cxn.Close()

//     log.Printf("Server Conn: Connection received from remote addr: %v, remote port: %d\n",
//         cxn.sockaddr.Addr, cxn.sockaddr.Port)

//     // Open the video file
//     file, err := os.Open("/home/rohan/www.example.org/1.mkv")
//     if err != nil {
//         log.Println("Failed to open video file:", err)
//         return
//     }
//     defer file.Close()

//     // Copy the contents of the video file to the connection
//     _, err = io.Copy(cxn, file)
//     if err != nil {
//         log.Println("Failed to send video file:", err)
//         return
//     }

//     log.Println("Video file sent successfully")
// }

// // Gracefully close the connection to a client
// func (cxn *TFOServerConn) Close() {

// 	// Gracefull close the connection
// 	err := syscall.Shutdown(cxn.fd, syscall.SHUT_RDWR)
// 	if err != nil {
// 		log.Println("Failed to shutdown() connection:", err)
// 	}

// 	// Close the file descriptor
// 	err = syscall.Close(cxn.fd)
// 	if err != nil {
// 		log.Println("Failed to close() connection:", err)
// 	}

// }


// func main(){
// 	var serverAddr [4]byte

// 	IP := net.ParseIP("127.0.0.1")
// 	if IP == nil {
// 		log.Fatal("Unable to process IP: ", "127.0.0.1")
// 	}

// 	copy(serverAddr[:], IP[12:16])

// 	server := TFOServer{ServerAddr: serverAddr, ServerPort: 8080}
// 	err := server.Bind()
// 	if err != nil {
// 		log.Fatalln("Failed to bind socket:", err)
// 	}

// 	// Create a new routine ("thread") and wait for connection from client
// 	server.Accept()
	
// }



package main

import (
    "errors"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "syscall"
)

type TFOServer struct {
    ServerAddr [4]byte
    ServerPort int
    fd         int
}

const TCP_FASTOPEN int = 23
const LISTEN_BACKLOG int = 23

// Create a tcp socket, setting the TCP_FASTOPEN socket option.
func (s *TFOServer) Bind() (err error) {

    s.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
    if err != nil {
        if err == syscall.ENOPROTOOPT {
            err = errors.New("TCP Fast Open server support is unavailable (unsupported kernel).")
        }
        return
    }

    err = syscall.SetsockoptInt(s.fd, syscall.SOL_TCP, TCP_FASTOPEN, 1)
    if err != nil {
        err = errors.New(fmt.Sprintf("Failed to set necessary TCP_FASTOPEN socket option: %s", err))
        return
    }

    sa := &syscall.SockaddrInet4{Addr: s.ServerAddr, Port: s.ServerPort}

    err = syscall.Bind(s.fd, sa)
    if err != nil {
        err = errors.New(fmt.Sprintf("Failed to bind to Addr: %v, Port: %d, Reason: %s", s.ServerAddr, s.ServerPort, err))
        return
    }

    log.Printf("Server: Bound to addr: %v, port: %d\n", s.ServerAddr, s.ServerPort)

    err = syscall.Listen(s.fd, LISTEN_BACKLOG)
    if err != nil {
        err = errors.New(fmt.Sprintf("Failed to listen: %s", err))
        return
    }

    return

}

// Block, waiting for connections, handling each connection in its own go
// routine
func (s *TFOServer) Accept() {

    log.Println("Server: Waiting for connections")

    defer syscall.Close(s.fd)

    for {

        fd, sockaddr, err := syscall.Accept(s.fd)
        if err != nil {
            log.Fatalln("Failed to accept(): ", err)
        }

        cxn := TFOServerConn{fd: fd, sockaddr: sockaddr.(*syscall.SockaddrInet4)}

        go cxn.Handle()

    }

}

// A client/server connection accepted by TFOServer
type TFOServerConn struct {
    sockaddr *syscall.SockaddrInet4
    fd       int
}

func (cxn *TFOServerConn) Write(p []byte) (n int, err error) {
    return syscall.Write(cxn.fd, p)
}

func (cxn *TFOServerConn) Handle() {
    defer cxn.Close()

    log.Printf("Server Conn: Connection received from remote addr: %v, remote port: %d\n",
        cxn.sockaddr.Addr, cxn.sockaddr.Port)

    // Open the video file
    file, err := os.Open("/home/rohan/www.example.org/1.mkv")
    if err != nil {
        log.Println("Failed to open video file:", err)
        return
    }
    defer file.Close()

    // Copy the contents of the video file to the connection
    _, err = io.Copy(cxn, file)
    if err != nil {
        log.Println("Failed to send video file:", err)
        return
    }

    log.Println("Video file sent successfully")
}
// Gracefully close the connection to a client
func (cxn *TFOServerConn) Close() {

	// Gracefull close the connection
	err := syscall.Shutdown(cxn.fd, syscall.SHUT_RDWR)
	if err != nil {
		log.Println("Failed to shutdown() connection:", err)
	}

	// Close the file descriptor
	err = syscall.Close(cxn.fd)
	if err != nil {
		log.Println("Failed to close() connection:", err)
	}

}


func main(){
	var serverAddr [4]byte

	IP := net.ParseIP("127.0.0.1")
	if IP == nil {
		log.Fatal("Unable to process IP: ", "127.0.0.1")
	}

	copy(serverAddr[:], IP[12:16])

	server := TFOServer{ServerAddr: serverAddr, ServerPort: 8080}
	err := server.Bind()
	if err != nil {
		log.Fatalln("Failed to bind socket:", err)
	}

	// Create a new routine ("thread") and wait for connection from client
	server.Accept()
	
}