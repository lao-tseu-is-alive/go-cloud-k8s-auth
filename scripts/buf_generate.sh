#!/bin/bash
if [ ! -d "./gen" ]; then
  mkdir gen
fi

buf dep update
buf generate