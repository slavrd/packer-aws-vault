#!/usr/bin/env bash
# initializes vault and records the unseal keys and root token to files.

[ -z ${VAULT_ADDR} ] && export VAULT_ADDR="http://127.0.0.1:8200"
[ -z ${VAULT_UNSEAL_KEYS_PATH} ] && VAULT_UNSEAL_KEYS_PATH="/etc/vault.d/.unseal-keys"
[ -z ${VAULT_ROOT_TOKEN_PATH} ] && VAULT_ROOT_TOKEN_PATH="/etc/vault.d/.vault-token"

which vault >/dev/null || {
    echo "==> ERROR Vault is not installed" >&2
    exit 1
}

vault operator init >vault_init.log 2>&1 && {
    cat vault_init.log | grep 'Unseal Key' | sed -e 's/.*: \(.*\)/\1/' | sudo tee ${VAULT_UNSEAL_KEYS_PATH} >/dev/null
    cat vault_init.log | grep 'Initial Root Token:' | sed -e 's/.*: \(.*\)/\1/' | sudo tee ${VAULT_ROOT_TOKEN_PATH} >/dev/null
    sudo chmod 600 ${VAULT_UNSEAL_KEYS_PATH} ${VAULT_ROOT_TOKEN_PATH}
}

[ -f vault_init.log ] && rm -f vault_init.log
