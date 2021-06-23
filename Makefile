.PHONY: todosvc docker-image docker-push-image

todosvc:
	mkdir -p dist
	go build -o dist/todosvc todosvc.go

docker-image:
	docker build -t coopernurse/todosvc .
	docker tag coopernurse/todosvc docker.io/coopernurse/todosvc

docker-push-image:
	docker push docker.io/coopernurse/todosvc
