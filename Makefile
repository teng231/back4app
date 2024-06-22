build:
	go build
run:
	PORT=8000 \
	DOMAIN_ALLOWED="*" \
	./back4app start
start: build run