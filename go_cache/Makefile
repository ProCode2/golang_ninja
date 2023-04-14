build:
	go build -o bin/gocache

run: build
	./bin/gocache

runfollower: build
	./bin/gocache --listenaddr :8000 --leaderaddr :3000
