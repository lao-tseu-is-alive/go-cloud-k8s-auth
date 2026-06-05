#!/bin/bash
if [ ! -d "./gen" ]; then
  mkdir gen
fi

buf dep update api/proto
buf generate api/proto