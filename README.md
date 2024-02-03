# time-sql


go run g.go -u https://molkenmusic.com/store/shop/index.php?cat=dvds -p payloads.txt -v -o p.txt

go install github.com/tomnomnom/time-sql@latest

cp -r /root/go/bin/time-sql /usr/local/bin

time-sql -l sub.txt -t 5 -o url.txt

time-sql -u tesla.com -t 5 -o url.txt
