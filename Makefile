DBUSERNAME=root

run:
	docker build -t authenticate authenticate/
	docker build -t inventory inventory/
	docker build -t resources resources/
	docker build -t progress progress/
	docker swarm init
	docker stack deploy -c docker-compose.yml blueprint

clean:
	docker stack rm blueprint
	docker swarm leave --force

build:
	docker build -t authenticate authenticate/
	docker build -t inventory inventory/
	docker build -t resources resources/
	docker build -t progress progress/

test: test_auth test_inv test_res test_pro

test_auth:
	mysql -u $(DBUSERNAME) -p < database/create_test.sql
	docker build -f authenticate/Dockerfile_test -t authenticate_test authenticate/
	docker run authenticate_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql

test_inv:
	cat database/create_test.sql database/account_test.sql | mysql -u $(DBUSERNAME) -p
	docker build -f inventory/Dockerfile_test -t inventory_test inventory/
	docker run inventory_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql

test_res:
	cat database/create_test.sql database/account_test.sql | mysql -u $(DBUSERNAME) -p
	docker build -f resources/Dockerfile_test -t resources_test resources/
	docker run resources_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql

test_pro:
	cat database/create_test.sql database/account_test.sql | mysql -u $(DBUSERNAME) -p
	docker build -f progress/Dockerfile_test -t progress_test progress/
	docker run progress_test
	mysql -u $(DBUSERNAME) -p < database/drop_test.sql
