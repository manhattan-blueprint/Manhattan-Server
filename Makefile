DBUSERNAME=root
DBPASSWORD=

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

test: test_auth test_inv test_res

test_auth:
	mysql -u $(DBUSERNAME) $(DBPASSWORD) < database/create_test.sql
	docker build -f authenticate/Dockerfile_test -t authenticate_test authenticate/
	docker run authenticate_test
	mysql -u $(DBUSERNAME) $(DBPASSWORD) < database/drop_test.sql

test_inv:
	cat database/create_test.sql database/inventory_test.sql | mysql -u $(DBUSERNAME) $(DBPASSWORD)
	docker build -f inventory/Dockerfile_test -t inventory_test inventory/
	docker run inventory_test
	mysql -u $(DBUSERNAME) $(DBPASSWORD) < database/drop_test.sql

test_res:
	cat database/create_test.sql database/resources_test.sql | mysql -u $(DBUSERNAME) $(DBPASSWORD)
	docker build -f resources/Dockerfile_test -t resources_test resources/
	docker run resources_test
	mysql -u $(DBUSERNAME) $(DBPASSWORD) <database/drop_test.sql