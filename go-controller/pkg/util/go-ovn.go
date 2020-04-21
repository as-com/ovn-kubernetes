package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	goovn "github.com/ebay/go-ovn"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/config"
	"k8s.io/klog"
)

var OVNNBDBClient goovn.Client
var OVNSBDBClient goovn.Client

func InitOVNDBClients() error {
	var err error

	switch config.OvnNorth.Scheme {
	case config.OvnDBSchemeSSL:
		OVNNBDBClient, err = initGoOvnSslClient(config.OvnNorth.Cert,
			config.OvnNorth.PrivKey, config.OvnNorth.CACert,
			config.OvnNorth.GetURL(), goovn.DBNB)
	case config.OvnDBSchemeTCP:
		OVNNBDBClient, err = initGoOvnTcpClient(config.OvnNorth.GetURL(), goovn.DBNB)
	case config.OvnDBSchemeUnix:
		OVNNBDBClient, err = initGoOvnUnixClient(config.OvnNorth.GetURL(), goovn.DBNB)
	default:
		klog.Errorf("Invalid db scheme: %s when initializing the OVN NB Client",
			config.OvnNorth.Scheme)
	}

	if err != nil {
		return fmt.Errorf("Couldn't initialize NBDB client: %s", err)
	}

	klog.Infof("Created OVN NB client with Scheme: %s", config.OvnNorth.Scheme)

	switch config.OvnSouth.Scheme {
	case config.OvnDBSchemeSSL:
		OVNSBDBClient, err = initGoOvnSslClient(config.OvnSouth.Cert,
			config.OvnSouth.PrivKey, config.OvnSouth.CACert,
			config.OvnSouth.GetURL(), goovn.DBSB)
	case config.OvnDBSchemeTCP:
		OVNSBDBClient, err = initGoOvnTcpClient(config.OvnSouth.GetURL(), goovn.DBSB)
	case config.OvnDBSchemeUnix:
		OVNSBDBClient, err = initGoOvnUnixClient(config.OvnSouth.GetURL(), goovn.DBSB)
	default:
		klog.Errorf("Invalid db scheme: %s when initializing the OVN SB Client",
			config.OvnSouth.Scheme)
	}

	if err != nil {
		return fmt.Errorf("Couldn't initialize SBDB client: %s", err)
	}

	klog.Infof("Created OVN SB client with Scheme: %s", config.OvnSouth.Scheme)
	return nil
}

func initGoOvnSslClient(certFile, privKeyFile, caCertFile, address, db string) (goovn.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("Error generating x509 certs for ovndbapi: %s", err)
	}
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("Error generating ca certs for ovndbapi: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	ovndbclient, err := goovn.NewClient(&goovn.Config{
		Db:        db,
		Addr:      address,
		TLSConfig: tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating SSL OVNDBClient for database %s at address %s: %s", db, address, err)
	}
	klog.Infof("Created OVNDB SSL client for db: %s", db)
	return ovndbclient, nil
}

func initGoOvnTcpClient(address, db string) (goovn.Client, error) {
	ovndbclient, err := goovn.NewClient(&goovn.Config{
		Db:   db,
		Addr: address,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating TCP OVNDBClient for address %s: %s", address, err)
	}
	klog.Infof("Created OVNDB TCP client for db: %s", db)
	return ovndbclient, nil
}

func initGoOvnUnixClient(address, db string) (goovn.Client, error) {
	ovndbclient, err := goovn.NewClient(&goovn.Config{
		Db:   db,
		Addr: address,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating OVNDBClient for address %s: %s", address, err)
	}
	klog.Infof("Created OVNDB UNIX client for db: %s", db)
	return ovndbclient, nil
}
