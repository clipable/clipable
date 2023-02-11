all: prune clideps
	-docker volume create grafana_data minio_data 2> /dev/null
	docker run -d -p 9000:9000 -p 9001:9001 -v minio_data:/data --name=clipable-minio minio/minio server /data -console-address ":9001"
	docker run -d -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust --name clipable-postgres postgres:14
	docker run -d -p 9090:9090 --add-host=host.docker.internal:host-gateway -v ${PWD}/prometheus.yml:/etc/prometheus/prometheus.yml --name clipable-prometheus prom/prometheus:v2.1.0 --config.file=/etc/prometheus/prometheus.yml
	docker run -d -p 3001:3000 --add-host=host.docker.internal:host-gateway -v grafana_data:/var/lib/grafana --name clipable-grafana grafana/grafana

	until $$(docker run --network host --rm postgres:14 pg_isready -h localhost > /dev/null) ; do \
		printf '.' ; \
		sleep 1 ; \
	done

	# sqlboiler -c sqlboiler.toml --add-global-variants --wipe psql
	# swag init

prune:
	-docker rm -f clipable-postgres clipable-prometheus clipable-grafana clipable-minio 2> /dev/null

clideps:
	test -s ${GOPATH}/bin/sqlboiler || { go install github.com/volatiletech/sqlboiler/v4@latest; }
	test -s ${GOPATH}/bin/sqlboiler-psql || { go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest; }
	test -s ${GOPATH}/bin/swag || { go install github.com/swaggo/swag/cmd/swag@latest; }

up:
	docker start clipable-postgres clipable-prometheus clipable-grafana clipable-minio

down:
	docker stop clipable-postgres clipable-prometheus clipable-grafana clipable-minio