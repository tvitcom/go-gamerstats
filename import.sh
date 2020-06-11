#!/bin/sh
## run for impot into mongo db
export $(grep -v '^#' .env | xargs -d '\n')
#echo $db_name
mongoimport --host ${db_host} --port ${db_port} --db ${db_name} --collection users --drop --file ./users_go.json --batchSize 1
mongoimport --host ${db_host} --port ${db_port} --db ${db_name} --collection user_games --drop --file ./data/games.json --batchSize 1

