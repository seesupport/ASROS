#!/bin/bash
# Build the UEFI application
TINYGO=${TINYGO:-tinygo}
$TINYGO build -o rtos.efi -target=x86_64-unknown-uefi -no-debug ./boot/uefi
