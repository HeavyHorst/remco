#!/bin/bash

remco config file -c integration/file/file.toml
cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf