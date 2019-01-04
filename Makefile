DBUSERNAME=root

run:
	docker build -t authenticate authenticate/
	docker build -t inventory inventory/
	docker build -t resources resources/
	docker swarm init
	docker stack deploy -c docker-compose.yml blueprint

clean:
	docker stack rm blueprint
	docker swarm leave --force

build:
	docker build -t authenticate authenticate/
	docker build -t inventory inventory/
	docker build -t resources resources/

test_auth:
	mysql -u $(DBUSERNAME) -p < database/create_test.sql
	docker build -f authenticate/Dockerfile_test -t authenticate_test authenticate/
	docker run authenticate_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql

test_inv:
	mysql -u $(DBUSERNAME) -p < database/create_test.sql
	mysql -u $(DBUSERNAME) -p < database/inventory_test.sql
	docker build -f inventory/Dockerfile_test -t inventory_test inventory/
	docker run inventory_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql