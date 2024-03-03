package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	c := getConfig()
	assert.Equal(t, "gops_db", c.Database.Database)
	assert.Equal(t, 5432, c.Database.Port)

	// verify the account configs are read correctly
	assert.NotEmpty(t, c.Mail.Accounts)
	accounts := *c.Mail.Accounts
	assert.NotEmpty(t, accounts[0].EncUser)
	assert.NotEmpty(t, accounts[0].Prune.Folders)
}
