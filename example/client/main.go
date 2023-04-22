package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	// "net"

	"net"
	// "time"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/internal/testdata"
	"github.com/quic-go/quic-go/internal/utils"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
)

func tfoControl(network, address string, c syscall.RawConn) error {
    return c.Control(func(fd uintptr) {
        // Set TCP Fast Open option
        syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, 0x17, 1)
    })
}
type Request struct {
	Method     string
	Path       string
	Host       string
	RangeStart int
	RangeEnd   int
}
func main() {
    verbose := flag.Bool("v", false, "verbose")
    quiet := flag.Bool("q", false, "don't print the data")
    keyLogFile := flag.String("keylog", "", "key log file")
    insecure := flag.Bool("insecure", false, "skip certificate verification")
    enableQlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
    flag.Parse()
    urls := flag.Args()

    logger := utils.DefaultLogger

    if *verbose {
        logger.SetLogLevel(utils.LogLevelDebug)
    } else {
        logger.SetLogLevel(utils.LogLevelInfo)
    }
    logger.SetLogTimeFormat("")

    var keyLog io.Writer
    if len(*keyLogFile) > 0 {
        f, err := os.Create(*keyLogFile)
        if err != nil {
            log.Fatal(err)
        }
        defer f.Close()
        keyLog = f
    }

    pool, err := x509.SystemCertPool()
    if err != nil {
        log.Fatal(err)
    }
    testdata.AddRootCA(pool)

    var qconf quic.Config
    if *enableQlog {
        qconf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
            filename := fmt.Sprintf("client_%x.qlog", connID)
            f, err := os.Create(filename)
            if err != nil {
                log.Fatal(err)
            }
            log.Printf("Creating qlog file %s.\n", filename)
            return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
        })
    }
    roundTripper := &http3.RoundTripper{
        TLSClientConfig: &tls.Config{
            RootCAs:            pool,
            InsecureSkipVerify: *insecure,
            KeyLogWriter:       keyLog,
        },
        QuicConfig: &qconf,
    }
    defer roundTripper.Close()
	// tr := &http.Transport{
    //     MaxIdleConns:        10,
    //     IdleConnTimeout:     30,
    //     DisableCompression:  true,
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    // }
	// client := &http.Client{Transport: tr}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification (for testing only!)
		},
		DialContext: (&net.Dialer{
			Control: func(network, address string, c syscall.RawConn) error {
				var err error
				err = c.Control(func(fd uintptr) {
					_ = syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, 0x17, 1)
				})
				return err
			},
		}).DialContext,
	}

	client := &http.Client{Transport: transport}

    hclient := &http.Client{
        Transport: roundTripper,
    }

    var wg sync.WaitGroup
    wg.Add(len(urls))
    for _, addr := range urls {
        logger.Infof("GET %s", addr)
        go func(addr string) {
            rsp, err := hclient.Get(addr)
            if err != nil {
                log.Fatal(err)
            }
            logger.Infof("Got response for %s: %#v", addr, rsp)

            body := &bytes.Buffer{}
            sigCh := make(chan os.Signal, 1)
            signal.Notify(sigCh, syscall.SIGINT)
            copyErrCh := make(chan error, 1)
            go func() {
                _, err := io.Copy(body, rsp.Body)
                if err != nil {
                    logger.Infof("response body %d",body.Len())
                    log.Fatal(err)
                }
                copyErrCh <- err
            }()
            select {
            case <-copyErrCh:
                // The copy completed without errors
                fmt.Printf("Copied %d bytes\n", body.Len())
            case <-sigCh:
                // Ctrl+C was pressed, so cancel the copy and exit gracefully
                fmt.Println("Interrupted")
                logger.Infof("bytes read so far %d",body.Len())
                // I will make a TCP connection here so that further data can be fetched from here
                
                bytesCopied := int64(body.Len())

                // create new HTTP request to fetch remaining bytes
                req, err := http.NewRequest("GET", addr, nil)
                if err != nil {
                    log.Fatal(err)
                }
                req.Header.Set("Range", fmt.Sprintf("bytes=%d-", bytesCopied))
		
			

                // make HTTP request with new TCP connection
                // rsp, err := hclient.Do(req)
				rsp, err := client.Do(req)
                if err != nil {
                    log.Fatal(err)
                }

                logger.Infof("Got response for %s: %#v", addr, rsp)
                
                buf := &bytes.Buffer{}
                // read remaining bytes from response body
                _, err = io.Copy(buf, rsp.Body)
                if err != nil {
                    log.Fatal(err)
                }

                logger.Infof("response body after TCP %d",buf.Len())
               
                os.Exit(0)
            }
        
            // Print the number of bytes copied
            logger.Infof("response body %d",body.Len())
            if *quiet {
                logger.Infof("Response Body: %d bytes", body.Len())
            } else {
                logger.Infof("Response Body:")
                logger.Infof("%s", body.Bytes())
            }
            wg.Done()
        }(addr)
    }
    wg.Wait()
}