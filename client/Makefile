NAME       := cloudflare-ddns-client
SCRIPTNAME := $(NAME).py

DSTDIR  := /usr/local/bin
ENVFILE := /usr/local/etc/$(NAME).conf
USER    := root
GROUP   := root

.PHONY: install
install:
	@command cp $(SCRIPTNAME) $(DSTDIR)/
	@chown $(USER):$(GROUP) $(DSTDIR)/$(SCRIPTNAME)
	@echo -e "APITOKEN=" | tee $(ENVFILE)

.PHONY: uninstall
uninstall: revoke_service
	@rm -f $(DSTDIR)/$(SCRIPTNAME)
	@rm -f $(ENVFILE)

create_service:
	@echo -e "[Unit]\nDescription=$(NAME)\nWants=network-online.target\nAfter=network-online.target\n\n[Service]\nType=oneshot\nUser=$(USER)\nGroup=$(GROUP)\n\nEnvironmentFile=$(ENVFILE)\nWorkingDirectory=$(DSTDIR)/\nExecStart=/usr/bin/python3 $(SCRIPTNAME) \$${APITOKEN}\n\n[Install]\nWantedBy=network-online.target" | tee /etc/systemd/system/$(NAME).service
	@echo -e "[Unit]\nDescription=Runs $(NAME).service regularly\n\n[Timer]\nOnBootSec=5min\nOnUnitActiveSec=30min\nRandomizedDelaySec=3min\nUnit=$(NAME).service\n\n[Install]\nWantedBy=timers.target" | tee /etc/systemd/system/$(NAME).timer
	@systemctl enable $(NAME).service
	@systemctl enable $(NAME).timer

.PHONY: start
start: create_service
	@systemctl start $(NAME).service

revoke_service: /etc/systemd/system/$(NAME).service /etc/systemd/system/$(NAME).timer
	@systemctl stop $(NAME).timer
	@systemctl disable $(NAME).service
	@systemctl disable $(NAME).timer
	@rm -f /etc/systemd/system/$(NAME).service
	@rm -f /etc/systemd/system/$(NAME).timer
	@systemctl daemon-reload
