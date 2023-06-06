package ldapSearchClient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/inverse-inc/go-utils/log"
	"github.com/inverse-inc/packetfence/go/common/ldapClient"
	"github.com/inverse-inc/packetfence/go/pfconfigdriver"
	"gopkg.in/ldap.v2"
)

type SearchQuery struct {
	Server     string          `json:"server"`
	Filter     string          `json:"filter"`
	Attributes []string        `json:"attributes,omitempty"`
	Context    context.Context `json:"context"`
}

type LdapServer struct {
	pfconfigdriver.AuthenticationSourceLdap
}

type LdapSearchClient struct {
	LdapServer        *LdapServer
	Timeout           time.Duration
	LdapClientFactory ldapClient.ILdapClientFactory
}

var (
	scopes = map[string]int{
		"base": ldap.ScopeBaseObject,
		"one":  ldap.ScopeSingleLevel,
		"sub":  ldap.ScopeWholeSubtree,
	}
)

func (sc LdapSearchClient) SearchLdap(query *SearchQuery) (map[string]map[string]interface{}, error) {
	conn := sc.connect()
	if conn == nil {
		return nil, errors.New("failed to connect to the LDAP server")
	}

	defer conn.Close()

	scope, ok := scopes[sc.LdapServer.Scope]

	if !ok {
		return nil, errors.New("unknown search scope: " + sc.LdapServer.Scope)
	}

	response, err := conn.SearchWithPaging(&ldap.SearchRequest{
		BaseDN:     sc.LdapServer.BaseDN,
		Scope:      scope,
		Filter:     query.Filter,
		Attributes: query.Attributes,
	}, uint32(200))

	if err != nil {
		return nil, errors.New("failed search: " + err.Error())
	}

	_, err = json.MarshalIndent(transform(response.Entries), "", "  ")
	if err != nil {
		return nil, errors.New("cannot unmarshall: " + err.Error())
	}

	return transform(response.Entries), nil
}

func (sc LdapSearchClient) connect() ldapClient.ILdapConnection {
	ctx := log.LoggerNewContext(context.Background())
	sources := sc.LdapServer.Host
	sslEnabled := sc.LdapServer.Encryption == "ssl"
	tlsEnabled := sc.LdapServer.Encryption == "starttls"

	for _, src := range sources {
		serverSocketAddress := fmt.Sprintf("%s:%s", src, sc.LdapServer.Port)
		ldapClient := sc.LdapClientFactory.NewLdapClient("tcp", serverSocketAddress, sc.Timeout)
		conn, err := sc.checkConnection(serverSocketAddress, sslEnabled, ldapClient)

		if err != nil {
			log.LogInfo(ctx, "Failed to connect to the LDAP server: "+err.Error())
			continue
		}

		if tlsEnabled {
			//TODO this uses the exact same port. what about negotiation? we need 2 ports? 389 and 636? was this tested?
			err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		}

		if err != nil {
			log.LogInfo(ctx, "Failed to re-connect to an LDAP server using TLS: "+err.Error())
			conn.Close()
			continue
		}

		if err = conn.Bind(sc.LdapServer.BindDN, sc.LdapServer.Password); err != nil {
			log.LogInfo(ctx, "Failed to authenticate to an LDAP server: "+err.Error())
			conn.Close()
			continue
		}

		return conn
	}

	return nil
}

func (sc LdapSearchClient) checkConnection(serverSocketAddress string, sslEnabled bool, ldapClient ldapClient.ILdapClient) (conn ldapClient.ILdapConnection, err error) {
	if sslEnabled {
		conn, err = ldapClient.DialTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = ldapClient.Dial()
	}
	return conn, err
}

func transform(entries []*ldap.Entry) map[string]map[string]interface{} {
	res := make(map[string]map[string]interface{})
	for _, entry := range entries {
		res[entry.DN] = transformAttribs(entry.Attributes)
	}
	return res
}

func transformAttribs(attribs []*ldap.EntryAttribute) map[string]interface{} {
	res := make(map[string]interface{})
	for _, attrib := range attribs {
		switch len(attrib.Values) {
		case 0:
			res[attrib.Name] = nil
		case 1:
			res[attrib.Name] = attrib.Values[0]
		default:
			res[attrib.Name] = attrib.Values
		}
	}
	return res
}
