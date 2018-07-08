test: 	
	go test -coverprofile=coverage.out -covermode=count

test-race:
	go test -race

cover:
	go test -coverprofile=coverage.out -covermode=count

coverhtml:
	go tool cover -html=coverage.out