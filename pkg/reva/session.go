/*
 * MIT License
 *
 * Copyright (c) 2020 Daniel Mueller
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package reva

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	registry "github.com/cs3org/go-cs3apis/cs3/auth/registry/v1beta1"
	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
)

// Session stores information about a Reva session.
type Session struct {
	ctx    context.Context
	client gateway.GatewayAPIClient

	token string
}

func (session *Session) initSession(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("no context provided")
	}
	session.ctx = ctx

	return nil
}

// Initiate begins the actual session by creating a connection to the host and preparing the gateway client.
func (session *Session) Initiate(host string, insecure bool) error {
	if host == "" {
		return fmt.Errorf("no host provided")
	}

	// We first need to get a gRPC connection to the host
	if conn, err := session.getConnection(host, insecure); err == nil {
		// Create the gateway client
		session.client = gateway.NewGatewayAPIClient(conn)
		return nil
	} else {
		return fmt.Errorf("unable to establish a gRPC connection to '%v': %v", host, err)
	}
}

func (session *Session) getConnection(host string, insecure bool) (*grpc.ClientConn, error) {
	if insecure {
		return grpc.Dial(host, grpc.WithInsecure())
	} else {
		tlsconf := &tls.Config{InsecureSkipVerify: false}
		creds := credentials.NewTLS(tlsconf)
		return grpc.Dial(host, grpc.WithTransportCredentials(creds))
	}
}

// GetLoginMethods returns a list of all available login methods supported by the Reva instance.
func (session *Session) GetLoginMethods() ([]string, error) {
	req := &registry.ListAuthProvidersRequest{}
	if res, err := session.client.ListAuthProviders(session.ctx, req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return []string{}, err
		}

		methods := make([]string, 0, len(res.Types))
		for _, method := range res.Types {
			methods = append(methods, method)
		}
		return methods, nil
	} else {
		return []string{}, err
	}
}

// Login logs into Reva using the specified method and user credentials.
func (session *Session) Login(method string, username string, password string) error {
	req := &gateway.AuthenticateRequest{
		Type:         method,
		ClientId:     username,
		ClientSecret: password,
	}

	if res, err := session.client.Authenticate(session.ctx, req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return err
		}

		if res.Token == "" {
			return fmt.Errorf("invalid token received: %q", res.Token)
		}
		session.token = res.Token

		// Now that we have a valid token, we can append this to our context
		session.ctx = context.WithValue(session.ctx, common.AccessTokenIndex, session.token)
		session.ctx = metadata.AppendToOutgoingContext(session.ctx, common.AccessTokenName, session.token)

		return nil
	} else {
		return err
	}
}

// BasicLogin tries to log into Reva using basic authentication.
func (session *Session) BasicLogin(username string, password string) error {
	// Check if the 'basic' method is actually supported by the Reva instance; only continue if this is the case
	if supportedMethods, err := session.GetLoginMethods(); err == nil {
		if common.FindStringNoCase(supportedMethods, "basic") == -1 {
			return fmt.Errorf("'basic' login method is not supported")
		}

		return session.Login("basic", username, password)
	} else {
		return fmt.Errorf("unable to get a list of all supported login methods: %v", err)
	}
}

// NewReadRequest returns an HTTP read request helper instance.
func (session *Session) NewReadRequest(endpoint string, transportToken string) (*HTTPRequest, error) {
	return newHTTPRequest(session, endpoint, "GET", transportToken, nil)
}

// NewWriteRequest returns an HTTP write request helper instance.
func (session *Session) NewWriteRequest(endpoint string, transportToken string, data io.Reader) (*HTTPRequest, error) {
	return newHTTPRequest(session, endpoint, "PUT", transportToken, data)
}

// Client gets the gateway client instance.
func (session *Session) Client() gateway.GatewayAPIClient {
	return session.client
}

// Token gets the session token.
func (session *Session) Context() context.Context {
	return session.ctx
}

// Token returns the current session token.
func (session *Session) Token() string {
	return session.token
}

// IsValid checks whether the session has been initialized and fully established.
func (session *Session) IsValid() bool {
	return session.client != nil && session.ctx != nil && session.token != ""
}

// NewSessionWithContext creates a new Reva session using the provided context.
func NewSessionWithContext(ctx context.Context) (*Session, error) {
	session := &Session{}
	if err := session.initSession(ctx); err != nil {
		return nil, fmt.Errorf("unable to initialize the session: %v", err)
	}

	return session, nil
}

// NewSession creates a new Reva session using a default background context.
func NewSession() (*Session, error) {
	return NewSessionWithContext(context.Background())
}
