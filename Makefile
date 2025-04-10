CLIENTS ?= 2

dev:
	@for /L %%i in (1,1,$(CLIENTS)) do ( \
		$(MAKE) -C cmd/client dev \
		)
	$(MAKE) -C cmd/server dev
