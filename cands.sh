#!/bin/bash

set -e

./compile.sh
./send.sh $@
