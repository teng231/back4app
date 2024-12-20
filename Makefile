build:
	go build
run:
	PORT=8000 \
	DOMAIN_ALLOWED="*" \
	DB_DSN="24HE6sB8xWPNKJB.root:2maTnO1Q0c67ROe7@tcp(gateway01.ap-southeast-1.prod.aws.tidbcloud.com:4000)/proj2?charset=utf8mb4&parseTime=True&loc=Local&tls=true" \
	DB_COMIC_DSN="root:123456@tcp(localhost:3306)/comic" \
	BOT_TOKEN="1195205858:AAESJuvk-M6CSS5U_MX1JgfGLiw9fqWSsp8" ./back4app start

syncdb: build
	DB_DSN="24HE6sB8xWPNKJB.root:2maTnO1Q0c67ROe7@tcp(gateway01.ap-southeast-1.prod.aws.tidbcloud.com:4000)/proj2?charset=utf8mb4&parseTime=True&loc=Local&tls=true" ./back4app syncdb
start: build run
crawler: build
	DB_COMIC_DSN="root:123456@tcp(localhost:3306)/comic" ./back4app crawler

synccomicdb: build
	DB_COMIC_DSN="root:123456@tcp(localhost:3306)/comic" ./back4app synccomicdb