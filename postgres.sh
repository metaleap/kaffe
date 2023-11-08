sudo rm -rf /home/_/.cache/.yodb_kaffe
mkdir /home/_/.cache/.yodb_kaffe
podman rm -f yodb_kaffe
podman run -p 5432:5432 -v /home/_/.cache/.yodb_kaffe:/var/lib/postgresql/data --name yodb_kaffe -e POSTGRES_USER=yodb_kaffe -e POSTGRES_PASSWORD=yodb_kaffe -e POSTGRES_DB=yodb_kaffe docker.io/library/postgres:latest
podman rm -f yodb_kaffe
