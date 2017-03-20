VERSION := `git describe --tags | cut -d '-' -f 1`.`git rev-parse --short HEAD`
APP := seed
RUNUSER ?= root

: gobin
	
gobin:
	go build -a -ldflags "-X main.version=$(VERSION)" -o $(APP).$(VERSION)
	echo $(VERSION) > ./VERSION

deploy-bucket:
	gsutil cp bucket/installer gs://demo-compute/

deploy:
	echo ex. user@127.0.0.1:/dir
	@read -r -p "> " DIST;\
		if [ -n "$$DIST" ] ; then \
			rsync -avP ./Makefile \
						./service.temp \
						./VERSION \
						./public \
						./templates \
						./scripts \
						./$(APP).`cat ./VERSION` $$DIST;\
		fi

start:
	systemctl start $(APP)

stop:
	systemctl stop $(APP)

restart:
	systemctl restart $(APP)

install: link

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

link:
	ln -sf ./$(APP).`cat ./VERSION` ./seed

upgrade: link
	systemctl restart $(APP)
	systemctl status $(APP)

clear:
	ls seed.* | grep -v `cat ./VERSION` | xargs rm

