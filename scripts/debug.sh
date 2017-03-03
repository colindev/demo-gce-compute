#!/bin/bash

cat > /tmp/debug << _EOF_
PWD=$(pwd)
USER=${USER}
HOME_DIR=$(ls /home)
_EOF_
