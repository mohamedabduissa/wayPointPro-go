# wayPointPro-go

wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc

git clone https://github.com/mohamedabduissa/wayPointPro-go.git

cd wayPointPro-go

go mod tidy

cat <<EOF > /root/wayPointPro-go/.env
DB_HOST=localhost
DB_PORT=5432
DB_USER=waypointpro_user
DB_PASSWORD=Dh3hMMjzhaLq5VL7RT
DB_NAME=waypointpro_1
OSRM_HOST=http://localhost:5000
PORT=8080
EOF

go run cmd/main.go

