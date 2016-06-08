#!/bin/bash

while true
do
	#echo "hit"
	image_process_id=$(pidof imageserver)
	if [[ -z $image_process_id ]]; then
		nohup /home/ubuntu/cs654/imageserver &
	else
		response=$(python availclient.py test1.jpg)
		if [[ $response == \{\"error\":* ]]; then
			echo "restarting"
			kill -9 $image_process_id
			nohup /home/ubuntu/cs654/imageserver &
		fi
	fi
	sleep 5
done 
