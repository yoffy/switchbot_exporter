switchbot_exporter: *.go
	go build

.PHONY: clean
clean:
	$(RM) switchbot_exporter

.PHONY: install
install: switchbot_exporter
	install -m 755 switchbot_exporter /usr/local/bin/
	install -m 644 switchbot_exporter.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl restart switchbot_exporter.service
	systemctl enable switchbot_exporter.service

.PHONY: uninstall
uninstall:
	systemctl disable switchbot_exporter.service
	systemctl stop switchbot_exporter.service
	$(RM) /usr/local/bin/switchbot_exporter
	$(RM) /etc/systemd/system/switchbot_exporter.service
	systemctl daemon-reload
