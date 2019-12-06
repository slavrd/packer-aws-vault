#!/usr/bin/env bash
# unseals Vault by using a file containing the keys

[ -z ${VAULT_ADDR} ] && export VAULT_ADDR="http://127.0.0.1:8200"
[ -z ${VAULT_UNSEAL_KEYS_PATH} ] && VAULT_UNSEAL_KEYS_PATH="/etc/vault.d/.unseal-keys"

[ -f ${VAULT_UNSEAL_KEYS_PATH} ] && {
    sudo -E head -n 3 ${VAULT_UNSEAL_KEYS_PATH} | xargs -L 1 vault operator unseal || {
        echo "==> ERROR unsealing vault" >&2
        exit 1
    }
} || {
    echo "==> ERROR key file ${VAULT_UNSEAL_KEYS_PATH} does not exist" >&2
    exit 1
}
