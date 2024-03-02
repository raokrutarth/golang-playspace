package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	InitConfig()
	assert.Equal(t, "archive", GlobalConfig.App.Name)
	assert.Equal(t, "irl_archive_dev", GlobalConfig.Database.Database)
	assert.Equal(t, 32100, GlobalConfig.Database.Port)

	// verify the account configs are read correctly
	assert.NotEmpty(t, GlobalConfig.Mail.Accounts)
	accounts := *GlobalConfig.Mail.Accounts
	assert.NotEmpty(t, accounts[0].EncUser)
	assert.NotEmpty(t, accounts[0].Prune.Folders)
}
