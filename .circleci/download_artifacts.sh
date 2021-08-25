#!/bin/bash

export CIRCLE_TOKEN='93aadf2ec5b369f635035d4fb7620f517d3f109d'

curl https://circleci.com/api/v1.1/project/github/arcology/eshing-engine/latest/artifacts?circle-token=$CIRCLE_TOKEN \
   | grep -o 'https://[^"]*' \
   | sed -e "s/$/?circle-token=$CIRCLE_TOKEN/" \
   | wget -O libeshing.so -v -i -