/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
)

var Conf VaultConfiguration = VaultConfiguration{
	Config: NewVault(),
}

func NewVault() vaultapi.Config {
	config := vaultapi.DefaultConfig()

	return *config
}
