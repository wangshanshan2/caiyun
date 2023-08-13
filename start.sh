cd blockchain-explorer/app/persistence/fabric/postgreSQL/db
sudo -u postgres ./createdb.sh
cd ../../../..
cd platform/fabric
sudo docker-compose up -d
cd ../../..
cd client
npm start
