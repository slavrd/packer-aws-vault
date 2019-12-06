#!/usr/bin/env bash
# Performs box final cleanup after provisioning.
# Intended to be run as a last provisioner

# clean up installations
sudo apt-get clean
sudo rm -rf /tmp/* 

# Removing leftover leases and persistent rules
echo "cleaning up dhcp leases"
rm /var/lib/dhcp/*

# Zero out the free space to save space in the final image:
echo "Zeroing device to make space..."
dd if=/dev/zero of=/EMPTY bs=1M
rm -f /EMPTY
