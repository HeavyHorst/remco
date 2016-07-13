#!/bin/bash

remco advanced --config integration/file/file.toml
cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf