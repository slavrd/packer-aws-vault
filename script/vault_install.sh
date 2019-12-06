#!/usr/bin/env bash
# installs Vault

# varify/install prerequisite packages
PKGS="jq unzip"
which ${PKGS} || {
    sudo apt-get update
    sudo apt-get install -y ${PKGS} || exit 1
}

# set $VAULT_VER to the lates if not already set
[ -z ${VAULT_VER} ] && {
    VAULT_VER=$(curl -sL https://releases.hashicorp.com/vault/index.json | jq -r '.versions[].version' | sort -V | egrep -v 'ent|beta|rc|alpha' | tail -n1)
}

# download Vault
FILE="vault_${VAULT_VER}_linux_amd64.zip"

curl -sSfO "https://releases.hashicorp.com/vault/${VAULT_VER}/${FILE}" || {
    echo "==> ERROR downloading vault" >&2
    exit 1
}

# install Vault
unzip ${FILE}
sudo chown root:root vault
sudo mv vault /usr/local/bin/
vault --version || {
    echo "==> ERROR installing vault" >&2
    exit 1
}

# set-up Vault aouto-complete for current user
vault -autocomplete-install
complete -C /usr/local/bin/vault vault

# clean up
rm -f ${FILE}
