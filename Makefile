VERSION := `git describe --tags`.`git rev-parse --short HEAD`
APP := seed
RUNUSER ?= root

: gobin
	
gobin:
	go build -a -ldflags "-X main.version=$(VERSION)" -o $(APP).$(VERSION)
	echo $(VERSION) > ./VERSION

deploy:
	echo ex. user@127.0.0.1:/dir
	@read -r -p "> " DIST;\
		if [ -n "$$DIST" ] ; then \
			rsync -avP ./Makefile \
						./service.temp \
						./VERSION \
						./views \
						./scripts \
						./$(APP).`cat ./VERSION` $$DIST;\
		fi

install: upgrade

	cat ./service.temp | sed 's@{PWD}@'`pwd`'@g' | sed 's/{USER}/'$(RUNUSER)'/g' > /etc/systemd/system/$(APP).service

	systemctl daemon-reload
	systemctl enable $(APP)
	systemctl start $(APP)
	systemctl status $(APP)

uninstall:
	systemctl stop $(APP)
	systemctl disable $(APP)

	rm -i /etc/systemd/system/$(APP).service
	systemctl daemon-reload

upgrade:
	ln -sf ./$(APP).`cat ./VERSION` ./seed

