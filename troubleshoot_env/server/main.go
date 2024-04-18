package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"github.com/gin-gonic/gin"
)

func getTlsConfigs(caCertPath string) (*tls.Config, error) {
	tlsConf := &tls.Config{}

	(*tlsConf).ClientAuth = tls.RequireAndVerifyClientCert

	//CA cert
	caCertContent, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read root certificate file: %s", err.Error()))
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertContent)
	if !ok {
		return nil, errors.New("Failed to parse root certificate authority")
	}
	(*tlsConf).ClientCAs = roots

	return tlsConf, nil
}

func server(server string, port int64, caCertPath string, serverCertPath string, serverKeyPath string) error {
	tlsConfs, tlsErr := getTlsConfigs(caCertPath)
	if tlsErr != nil {
		return tlsErr
	}
	
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		fmt.Printf("Host is: %s\n", c.Request.Host)
		c.JSON(http.StatusOK, gin.H{
			"server": server,
		})
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: router,
		TLSConfig: tlsConfs,
	}
	go func() {
		err := srv.ListenAndServeTLS(serverCertPath, serverKeyPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}()
	fmt.Printf("Started server %s\n", server)
	return nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	err := server("server1", 8443, "../credentials/output/ca.crt", "../credentials/output/server_client.crt", "../credentials/output/server_client.key")
    if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for {
        time.Sleep(1 * time.Hour)
    }
}