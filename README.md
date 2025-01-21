# wayPointPro-go

wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc

git clone https://github.com/mohamedabduissa/wayPointPro-go.git

cd wayPointPro-go

go mod tidy

go run cmd/main.go

