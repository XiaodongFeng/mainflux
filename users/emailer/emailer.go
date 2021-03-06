// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package emailer

import (
	"fmt"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/users"
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL string
	agent    *email.Agent
}

// New creates new emailer utility
func New(url string, c *email.Config) (users.Emailer, error) {
	e, err := email.New(c)
	if err != nil {
		return nil, err
	}
	return &emailer{resetURL: url, agent: e}, nil
}

func (e *emailer) SendPasswordReset(To []string, host string, token string) errors.Error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	content := fmt.Sprintf("%s", url)
	return e.agent.Send(To, "", "Password reset", "", content, "")
}
