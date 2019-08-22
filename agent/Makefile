# -------------------------------------------------
#  Setup
# -------------------------------------------------
COMPOSE_FILE=./docker/docker-compose.yml

# call one time to generate .env file with some pre-defined values
setup:
	$(shell echo "IMAGE_NAME=agent_app_container" > .env)
	$(shell echo "MY_ENV_USER=" >> .env)
	$(shell echo "MY_ENV_TOKEN=" >> .env)
	$(shell echo "MY_ENV_PRESET=DMELAB" >> .env)

# -------------------------------------------------
#  Local
# -------------------------------------------------
build:
	pushd cmd/webapp &&\
	go build &&\
	popd
	
start:
	pushd cmd/webapp &&\
	./webapp &&\
	popd
	
# -------------------------------------------------
#  Docker
# -------------------------------------------------
docker-build:
	docker-compose -f $(COMPOSE_FILE) build app
	
docker-up:
	docker-compose -f $(COMPOSE_FILE) up -d app