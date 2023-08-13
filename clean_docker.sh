sudo docker rm -f $(sudo docker ps -aq)
sudo docker network prune
sudo docker volume prune
cd fixtures && sudo docker-compose up -d
cd ..
#google-chrome index.html
sudo rm caiyun
go build
sudo ./caiyun
#sudo ./Blockchain-medical-imaging > output.log 2>&1 &
# cd explorer && sudo docker-compose up -d
