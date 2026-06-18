#!/bin/bash
TINYGO=${TINYGO:-tinygo}
# Generate ISR stubs
python3 tools/gen_isr.py
# Build
$TINYGO build -o rtos.efi -target=x86_64-unknown-uefi -no-debug ./boot/uefi
