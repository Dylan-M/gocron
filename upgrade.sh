#!/bin/bash
cd $(dirname $0)

sudo git pull
sudo git status
read -p "Press [Enter] key to upgrade gocron on this branch" 
sudo service gocron stop
sudo rm /usr/local/bin/gocron
sudo cp ./bin/gocron /usr/local/bin/
sudo service gocron start
sudo service gocron status