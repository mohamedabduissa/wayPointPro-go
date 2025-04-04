# wayPointPro-go
sudo apt update && sudo apt upgrade -y

wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc

source ~/.bashrc

git clone https://github.com/mohamedabduissa/wayPointPro-go.git

cd wayPointPro-go

go mod tidy

echo "DB_HOST=10.0.0.4
DB_PORT=5432
DB_USER=waypointpro_user
DB_PASSWORD=Dh3hMMjzhaLq5VL7RT
DB_NAME=waypointpro_1
OSRM_HOST=http://10.0.0.2:5000
VALHALLA_HOST=http://10.0.0.2:8002
PORT=8080
REDIS=10.0.0.4:6379
PLATFORM=VALHALLA" > .env

go build -o waypointpro cmd/main.go

sudo nano /etc/systemd/system/waypointpro.service

[Unit]
Description=WayPointPro Go Service
After=network.target

[Service]
User=root
WorkingDirectory=/root/wayPointPro-go
ExecStart=/root/wayPointPro-go/waypointpro
Restart=always
EnvironmentFile=/root/wayPointPro-go/.env
StandardOutput=append:/var/log/waypointpro.log
StandardError=append:/var/log/waypointpro.log

[Install]
WantedBy=multi-user.target


sudo systemctl daemon-reload

sudo systemctl enable waypointpro

sudo systemctl restart waypointpro

sudo systemctl status waypointpro

journalctl -u waypointpro.service --since "1 hour ago" --no-pager
