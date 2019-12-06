#!/usr/bin/env bash
# setup vault as service, using local backend

[ -z ${VAULT_SVC_PATH} ] && VAULT_SVC_PATH="/tmp/vault.service"
[ -z ${VAULT_CFG_PATH} ] && VAULT_CFG_PATH="/tmp/vault.hcl"

which vault || {
    echo "==> ERROR Vault is not installed" >&2
    exit 1
}

# give Vault the ability to use the mlock syscall without running the process as root.
sudo setcap cap_ipc_lock=+ep /usr/local/bin/vault

# create a user for vault
sudo useradd --system --home /etc/vault.d --shell /bin/false vault

# setup Vault as service
sudo mkdir -p /etc/vault.d
sudo cp  ${VAULT_CFG_PATH} /etc/vault.d/vault.hcl
sudo chown --recursive vault:vault /etc/vault.d
sudo chmod 640 /etc/vault.d/vault.hcl

sudo mkdir -p /var/vault
sudo chown --recursive vault:vault /var/vault

sudo cp ${VAULT_SVC_PATH} /etc/systemd/system/vault.service
sudo systemctl daemon-reload
sudo systemctl enable vault.service
sudo systemctl start vault.service
