docker build -t forum .
docker image prune
docker run -d --name forum-container -p 8080:8080 forum