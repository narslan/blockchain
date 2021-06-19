package http

import (
	"context"
	"fmt"

	"os"
	"path"
	"strconv"

	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/narslan/ethos"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"golang.org/x/sync/errgroup"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

// Server represents an HTTP server. It is meant to wrap all HTTP functionality
// used by the application so that dependent packages (such as cmd/*) do not
// need to reference the "net/http" package at all.
type Server struct {
	//router *gin.Engine
	srv *http.Server

	// Bind address & domain for the server's listener.
	// If domain is specified, server is run on TLS using acme/autocert.
	Addr   string
	Domain string
	Cert   string
	Key    string
	UseTLS bool

	BlockChain *ethos.BlockChain
}

// NewServer returns a new instance of Server.
func NewServer() *Server {

	srv := &http.Server{}
	s := &Server{srv: srv}
	r := gin.New()

	//start a new blockchain instance
	s.BlockChain = ethos.NewBlockChain()

	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return path.Base(file) + ":" + strconv.Itoa(line)
	}

	writer := zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	}
	writer.FormatFieldName = func(i interface{}) string { return fmt.Sprintf("%s=>", i) }

	log.Logger = log.Output(
		writer,
	).With().Logger()

	r.Use(logger.SetLogger(), gin.Recovery())
	r.Use(initCors())

	r.GET("/chain", s.Chain)
	r.GET("/mine", s.Mine)
	r.POST("/transaction/new", s.NewTx)
	srv.Handler = r
	return s
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {

	ctx := context.Background()
	s.srv.Addr = s.Addr
	// Open a listener on our bind address.
	if s.UseTLS {

		if s.Cert == "" || s.Key == "" {
			return ethos.Errorf(ethos.EINVALID, "certificate and key must be supplied to be able to use tls")
		}

		g, _ := errgroup.WithContext(ctx)

		g.Go(func() error {
			return s.srv.ListenAndServeTLS(s.Cert, s.Key)
		})

		return g.Wait()

	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {

		return s.srv.ListenAndServe()
	})

	return g.Wait()

	// Begin serving requests on the listener. We use Serve() instead of

}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func initCors() gin.HandlerFunc {
	config := cors.Config{
		AllowWildcard:   true,
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Access-Control-Allow-Origin", "Access-Control-Allow-Methods"},
	}

	err := config.Validate()
	if err != nil {
		fmt.Println(err.Error())
	}
	return cors.New(config)
}
