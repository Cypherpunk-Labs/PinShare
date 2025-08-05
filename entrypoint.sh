#!/bin/bash 
/usr/local/bin/start_ipfs daemon --migrate=true --agent-version-suffix=docker &
sleep 60
/opt/pinshare/bin/pinshare
