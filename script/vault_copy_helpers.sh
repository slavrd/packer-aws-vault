#!/usr/bin/env bash
# copies vault helpers from tmp

sudo mkdir -p /etc/vault.d/scripts/

sudo mv /tmp/vault_init.sh /etc/vault.d/scripts/vault_init.sh
sudo mv /tmp/vault_unseal.sh /etc/vault.d/scripts/vault_unseal.sh
