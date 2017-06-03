#!/bin/bash -e

echo "==> Updating consul backup"
echo "--> Checksuming previous consul backup"
SumBefore=$(md5sum consul-backup.json | cut -f 1 -d ' ')
DateBefore=$(jq '.date' -r < consul-backup.json)

echo "--> Taking new consul backup"
consul-kv-backup backup > consul-full-backup.json
cp consul-full-backup.json consul-backup.json

echo "    Redacting secrets from backup"

BadVal=$(jq '.values["aws-creds/access_key_id"]' -r < consul-backup.json)
sed -i "s#$BadVal#REDACTED#" consul-backup.json

BadVal=$(jq '.values["aws-creds/secret_access_key"]' -r < consul-backup.json)
sed -i "s#$BadVal#REDACTED#" consul-backup.json

BadVal=$(jq '.values["aws-github-queue/queue-url"]' -r < consul-backup.json)
sed -i "s#$BadVal#https://sqs.us-west-2.amazonaws.com/REDACTED/sd-github-inbound#" consul-backup.json

BadVal=$(jq '.values["hue-bridge/username"]' -r < consul-backup.json)
sed -i "s#$BadVal#REDACTED#" consul-backup.json

echo "--> Checking backup for changes"
DateAfter=$(jq '.date' -r < consul-backup.json)
sed -i "s#$DateAfter#$DateBefore#" consul-backup.json
SumAfter=$(md5sum consul-backup.json | cut -f 1 -d ' ')
if [ "$SumBefore" == "$SumAfter" ]; then
  echo "    Consul backup hasn't changed since" $DateBefore
else
  echo "==> Consul backup has changed. Setting date to" $DateAfter
  sed -i "s#$DateBefore#$DateAfter#" consul-backup.json
fi

echo "==> Reformatting go source"
go fmt ./...
