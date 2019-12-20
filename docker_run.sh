#!/bin/sh

# sync s3 bucket every few minutes
sync_poll() {
  while true
  do
    sleep $SAVE_INTERVAL
    if [ $USE_S3 = "true" ]
    then
    	echo "syncing with s3"
    	aws s3 sync $GRAPH_DB_STORE_DIR $AWS_SYNC_DIRECTORY 
    fi
  done
}

# download from s3
if [ $USE_S3 = "true" ]
then
	echo "configuring s3"
	aws configure set aws_access_key_id $AWS_ACCESS_KEY_ID
	aws configure set aws_secret_access_key $AWS_SECRET_ACCESS_KEY
	aws configure set default.region $AWS_DEFAULT_REGION
	aws s3 sync $AWS_KV_PATH $GRAPH_DB_STORE_DIR
fi

# start cron job
sync_poll &

# start app
printenv
echo "$GRAPH_DB_STORE_DIR : "
ls -lh $GRAPH_DB_STORE_DIR
twowaykv $COMMAND