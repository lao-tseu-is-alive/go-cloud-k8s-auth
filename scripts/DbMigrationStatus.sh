#!/bin/bash
echo "will check migratin status with dbmate : https://github.com/amacneil/dbmate"
cd cmd/goCloudAuthServer                                                                                    
dbmate --env-file ../../.env status

