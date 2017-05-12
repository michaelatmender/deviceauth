// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package tenant

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/deviceauth/client"
	"github.com/mendersoftware/deviceauth/utils"
)

const (
	// devices endpoint
	TenantVerifyUri = "/api/internal/v1/tenantadm/tenants/verify"
	// default request timeout, 10s?
	defaultReqTimeout = time.Duration(10) * time.Second
)

var (
	ErrTokenVerificationFailed = errors.New("token verification failed")
)

// ClientConfig conveys client configuration
type Config struct {
	// Tenant administrator service address
	TenantAdmAddr string
	// Request timeout
	Timeout time.Duration
}

// ClientRunner is an interface of inventory client
type ClientRunner interface {
	VerifyToken(ctx context.Context, token string, client client.HttpRunner) error
}

// Client is an opaque implementation of inventory client. Implements
// ClientRunner interface
type Client struct {
	conf Config
}

// VerifyToken will execute a request to tenenatadm's endpoint for token
// verification. Returns nil if verification was successful.
func (tc *Client) VerifyToken(ctx context.Context, token string, client client.HttpRunner) error {

	// TODO sanitize token

	req, err := http.NewRequest(http.MethodPost,
		utils.JoinURL(tc.conf.TenantAdmAddr, TenantVerifyUri),
		nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request to tenant administrator")
	}

	// tenant token is passed in Authorization header
	req.Header.Add("Authorization", "Bearer "+token)

	ctx, cancel := context.WithTimeout(ctx, tc.conf.Timeout)
	defer cancel()

	rsp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "request to verify token failed")
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {

	case http.StatusUnauthorized: // 401, verification result negative
		return ErrTokenVerificationFailed

	case http.StatusOK: // 200, token verified
		return nil
	default:
		return errors.Errorf("token verification request returned unexpected status %v",
			rsp.StatusCode)
	}
}

// NewClient creates a client with given config.
func NewClient(c Config) *Client {
	if c.Timeout == 0 {
		c.Timeout = defaultReqTimeout
	}

	return &Client{
		conf: c,
	}
}
